package plugin_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/lock"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

type PluginManager struct {
	m mapping.Map[string, plugin_entities.PluginLifetime]

	cluster *cluster.Cluster

	maxPluginPackageSize int64
	workingDirectory     string

	// mediaManager is used to manage media files like plugin icons, images, etc.
	mediaManager *media_manager.MediaManager

	// running plugin in storage contains relations between plugin packages and their running instances
	runningPluginInStorage mapping.Map[string, string]
	// start process lock
	startProcessLock *lock.HighGranularityLock
	// serverless runtime

	// Install is a function that installs a plugin to the platform
	Install func(decoder decoder.PluginDecoder) (*stream.Stream[PluginInstallResponse], error)
}

var (
	manager *PluginManager
)

func InitManager(cluster *cluster.Cluster, configuration *app.Config) {
	manager = &PluginManager{
		cluster:              cluster,
		maxPluginPackageSize: configuration.MaxPluginPackageSize,
		workingDirectory:     configuration.PluginWorkingPath,
		mediaManager: media_manager.NewMediaManager(
			configuration.PluginMediaCachePath,
			configuration.PluginMediaCacheSize,
		),
		startProcessLock: lock.NewHighGranularityLock(),
	}

	if configuration.Platform == app.PLATFORM_AWS_LAMBDA {
		manager.Install = manager.InstallToAWSFromPkg
		serverless.Init(configuration)
	} else if configuration.Platform == app.PLATFORM_LOCAL {
		manager.Install = manager.InstallToLocal
	}

	manager.Init(configuration)
}

func Manager() *PluginManager {
	return manager
}

func (p *PluginManager) Add(
	plugin plugin_entities.PluginLifetime,
) error {
	identity, err := plugin.Identity()
	if err != nil {
		return err
	}

	p.m.Store(identity.String(), plugin)
	return nil
}

func (p *PluginManager) Get(
	identity plugin_entities.PluginUniqueIdentifier,
) plugin_entities.PluginLifetime {
	if v, ok := p.m.Load(identity.String()); ok {
		return v
	}

	// check if plugin is a serverless runtime
	plugin_session_interface, err := p.getServerlessPluginRuntime(identity)
	if err != nil {
		return nil
	}

	return plugin_session_interface
}

func (p *PluginManager) GetAsset(id string) ([]byte, error) {
	return p.mediaManager.Get(id)
}

func (p *PluginManager) Init(configuration *app.Config) {
	// TODO: init plugin manager
	log.Info("start plugin manager daemon...")

	// init redis client
	if err := cache.InitRedisClient(
		fmt.Sprintf("%s:%d", configuration.RedisHost, configuration.RedisPort),
		configuration.RedisPass,
	); err != nil {
		log.Panic("init redis client failed: %s", err.Error())
	}

	if err := dify_invocation.InitDifyInvocationDaemon(
		configuration.PluginInnerApiURL, configuration.PluginInnerApiKey,
	); err != nil {
		log.Panic("init dify invocation daemon failed: %s", err.Error())
	}

	// start local watcher
	if configuration.Platform == app.PLATFORM_LOCAL {
		p.startLocalWatcher(configuration)
	}

	// start remote watcher
	p.startRemoteWatcher(configuration)
}
