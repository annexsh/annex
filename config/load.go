package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Config interface {
	*AllServices | *TestService | *EventService | *WorkflowProxyService
}

func Load[C Config](dst C) error {
	loaderCfg := aconfig.Config{
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
		FileFlag:           "config-file",
		FailOnFileNotFound: true,
	}
	loader := aconfig.LoaderFor(dst, loaderCfg)
	return loader.Load()
}
