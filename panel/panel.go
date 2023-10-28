package panel

import (
	"os"

	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/gin-gonic/gin"
	"github.com/rodrikv/openai_proxy/panel/tables"
	"github.com/rodrikv/openai_proxy/utils"
)

type Panel struct {
	ginEngine *gin.Engine
	engine    *engine.Engine
}

func (p *Panel) Init() {
	var configFileName string

	p.ginEngine = gin.New()
	p.engine = engine.Default()

	fInfo, err := os.Stat(utils.Getenv("CONFIG_FILE", "config.json"))

	if err != nil {
		configFileName = ""
	} else {
		if fInfo.IsDir() {
			panic(fInfo.Name() + " is directory not a file")
		}
		configFileName = fInfo.Name()
	}

	cfg := p.GetConfig(configFileName)

	p.engine.AddConfig(cfg)

	st := tables.NewSystemTable(p.engine.DefaultConnection(), cfg)

	if err := p.engine.AddGenerators(map[string]table.Generator{
		"users":     st.GetUserTable,
		"models":    st.GetModelTable,
		"endpoints": st.GetEndpointTable,
	}).Use(p.ginEngine); err != nil {
		panic(err)
	}
}

func (p *Panel) Run(port string) {
	p.Init()
	p.ginEngine.Static("/uploads", "./upload")
	p.ginEngine.Static("/static", "./assets")

	p.ginEngine.Run(":" + port)
}

func (p *Panel) Stop() {

}

func (p *Panel) GetConfig(configPath string) *config.Config {
	var cfg config.Config

	if configPath != "" {
		cfg = config.ReadFromJson(configPath)
	} else {
		cfg = config.Config{
			Databases: config.DatabaseList{
				"default": config.Database{
					File:   "admin.db",
					Driver: "sqlite",
				},
			},
			UrlPrefix: "admin",
			Store: config.Store{
				Path:   "./uploads",
				Prefix: "uploads",
			},
			Language: language.EN,
			Theme:    "adminlte",
			Debug:    true,
		}
	}

	return &cfg
}
