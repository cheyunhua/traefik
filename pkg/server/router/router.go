package router

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/middlewares/accesslog"
	"github.com/containous/traefik/pkg/middlewares/recovery"
	"github.com/containous/traefik/pkg/middlewares/tracing"
	"github.com/containous/traefik/pkg/responsemodifiers"
	"github.com/containous/traefik/pkg/rules"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/containous/traefik/pkg/server/middleware"
	"github.com/containous/traefik/pkg/server/service"
)

const (
	recoveryMiddlewareName = "traefik-internal-recovery"
)

// NewManager Creates a new Manager
func NewManager(conf *config.RuntimeConfiguration,
	serviceManager *service.Manager,
	middlewaresBuilder *middleware.Builder,
	modifierBuilder *responsemodifiers.Builder,
) *Manager {
	return &Manager{
		routerHandlers:     make(map[string]http.Handler),
		serviceManager:     serviceManager,
		middlewaresBuilder: middlewaresBuilder,
		modifierBuilder:    modifierBuilder,
		conf:               conf,
	}
}

// Manager A route/router manager
type Manager struct {
	routerHandlers     map[string]http.Handler
	serviceManager     *service.Manager
	middlewaresBuilder *middleware.Builder
	modifierBuilder    *responsemodifiers.Builder
	conf               *config.RuntimeConfiguration
}

func (m *Manager) getHTTPRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*config.RouterInfo {
	if m.conf != nil {
		return m.conf.GetRoutersByEntrypoints(ctx, entryPoints, tls)
	}

	return make(map[string]map[string]*config.RouterInfo)
}

// BuildHandlers Builds handler for all entry points
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string, tls bool) map[string]http.Handler {
	entryPointHandlers := make(map[string]http.Handler)

	for entryPointName, routers := range m.getHTTPRouters(rootCtx, entryPoints, tls) {
		entryPointName := entryPointName
		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

		handler, err := m.buildEntryPointHandler(ctx, routers)
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}

		handlerWithAccessLog, err := alice.New(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, log.EntryPointName, entryPointName, accesslog.AddOriginFields), nil
		}).Then(handler)
		if err != nil {
			log.FromContext(ctx).Error(err)
			entryPointHandlers[entryPointName] = handler
		} else {
			entryPointHandlers[entryPointName] = handlerWithAccessLog
		}
	}

	m.serviceManager.LaunchHealthCheck()

	return entryPointHandlers
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*config.RouterInfo) (http.Handler, error) {
	router, err := rules.NewRouter()
	if err != nil {
		return nil, err
	}

	for routerName, routerConfig := range configs {
		ctxRouter := log.With(internal.AddProviderInContext(ctx, routerName), log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctxRouter)

		handler, err := m.buildRouterHandler(ctxRouter, routerName, routerConfig)
		if err != nil {
			routerConfig.Err = err.Error()
			logger.Error(err)
			continue
		}

		err = router.AddRoute(routerConfig.Rule, routerConfig.Priority, handler)
		if err != nil {
			routerConfig.Err = err.Error()
			logger.Error(err)
			continue
		}
	}

	router.SortRoutes()

	chain := alice.New()
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return recovery.New(ctx, next, recoveryMiddlewareName)
	})

	return chain.Then(router)
}

func (m *Manager) buildRouterHandler(ctx context.Context, routerName string, routerConfig *config.RouterInfo) (http.Handler, error) {
	if handler, ok := m.routerHandlers[routerName]; ok {
		return handler, nil
	}

	handler, err := m.buildHTTPHandler(ctx, routerConfig, routerName)
	if err != nil {
		return nil, err
	}

	handlerWithAccessLog, err := alice.New(func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, accesslog.RouterName, routerName, nil), nil
	}).Then(handler)
	if err != nil {
		log.FromContext(ctx).Error(err)
		m.routerHandlers[routerName] = handler
	} else {
		m.routerHandlers[routerName] = handlerWithAccessLog
	}

	return m.routerHandlers[routerName], nil
}

func (m *Manager) buildHTTPHandler(ctx context.Context, router *config.RouterInfo, routerName string) (http.Handler, error) {
	qualifiedNames := make([]string, len(router.Middlewares))
	for i, name := range router.Middlewares {
		qualifiedNames[i] = internal.GetQualifiedName(ctx, name)
	}
	rm := m.modifierBuilder.Build(ctx, qualifiedNames)

	sHandler, err := m.serviceManager.BuildHTTP(ctx, router.Service, rm)
	if err != nil {
		return nil, err
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	tHandler := func(next http.Handler) (http.Handler, error) {
		return tracing.NewForwarder(ctx, routerName, router.Service, next), nil
	}

	return alice.New().Extend(*mHandler).Append(tHandler).Then(sHandler)
}
