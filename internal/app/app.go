package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	// "time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"tgbot/embed"
	"tgbot/internal/app/middleware"
	"tgbot/internal/conf"
	"tgbot/internal/op"

	backend "tgbot/internal/app/handles/backend"
	backend_admin "tgbot/internal/app/handles/backend/admin"
	backend_log "tgbot/internal/app/handles/backend/log"
	backend_system "tgbot/internal/app/handles/backend/system"
	backend_tg "tgbot/internal/app/handles/backend/tg"
	"tgbot/internal/app/handles/home"
	"tgbot/internal/app/handles/install"
)

func initTmplFunc(r *gin.Engine) {
	// Define template functions
	funcMap := template.FuncMap{
		"safe": func(str string) template.HTML {
			return template.HTML(str)
		},
		// Cache-busting token exposed as a function for templates
		"BuildCommit": func() string {
			return conf.BuildCommit
		},
		"HasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		//是子菜单或当前菜单
		"IsSubOrEq": func(base, menu string) bool {
			if base == menu {
				return true
			}
			endp := strings.Replace(base, menu, "", 1)
			endp = strings.TrimPrefix(endp, "/")
			return !strings.Contains(endp, "/")
		},
		"Contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"json": func(v interface{}) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return string(b)
		},
		"formatBytes": func(bytes int64) string {
			if bytes < 1024 {
				return fmt.Sprintf("%d B", bytes)
			}
			return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
		},
	}

	// Build template set with directory-aware names (e.g., "install/index.tmpl")
	// so that we can reference templates across multiple directories explicitly.
	tpl := template.New("").Delims("{[", "]}").Funcs(funcMap)

	for _, name := range embed.TemplatesAllNames("templates") {
		// Trim the leading "templates/" so template names are like "install/index.tmpl"
		short := strings.TrimPrefix(name, "templates/")
		content, err := embed.Templates.ReadFile(name)
		if err != nil {
			panic(err)
		}
		if _, err := tpl.New(short).Parse(string(content)); err != nil {
			panic(err)
		}
	}

	r.SetHTMLTemplate(tpl)
}

// 后台/backstage
func initRuoteAdmin(r *gin.Engine) {
	// fmt.Println("conf.Web.AdminPath:", conf.Web.AdminPath)
	backstage := r.Group(conf.Web.AdminPath)
	backstage.Use(middleware.CheckInstalled())
	backstage.GET("/login", backend.LoginPage)
	backstage.POST("/login", backend.PostLogin)
	backstage.GET("/logout", backend.LoginOut)

	backstage_admin := backstage.Group("")
	backstage_admin.Use(middleware.CheckInstalled(), middleware.AuthRequired())

	// 管理员
	// backstage_admin.GET("", backend.HomePage)
	backstage_admin.GET("", backend_admin.Home)
	backstage_admin.GET("/index", backend_admin.Home)

	backstage_admin.GET("/admin/index", backend_admin.Home)
	backstage_admin.GET("/admin/add", backend_admin.Add)
	backstage_admin.POST("/admin/add", backend_admin.PostAdd)
	backstage_admin.GET("/admin/list", backend_admin.List)
	backstage_admin.GET("/admin/details", backend_admin.Details)
	backstage_admin.GET("/admin/update", backend_admin.Update)
	backstage_admin.POST("/admin/delete", backend_admin.Delete)
	backstage_admin.POST("/admin/trigger_status", backend_admin.AdminTriggerStatus)

	// 日志审计
	backstage_admin.GET("/log", backend_log.Home)
	backstage_admin.GET("/log/list", backend_log.List)
	backstage_admin.GET("/log/settings", backend_log.Settings)
	backstage_admin.POST("/log/settings", backend_log.PostSettting)
	backstage_admin.GET("/log/clean", backend_log.Clean)
	backstage_admin.POST("/log/clean", backend_log.PostLogClean)
	backstage_admin.POST("/log/delete", backend_log.Delete)

	// 系统设置
	backstage_admin.GET("/system/settings", backend_system.Home)
	backstage_admin.POST("/system/settings/home", backend_system.PostHome)

	backstage_admin.GET("/system/settings/web", backend_system.Web)
	backstage_admin.POST("/system/settings/web", backend_system.PostWeb)

	backstage_admin.GET("/system/settings/profile", backend_system.Profile)
	backstage_admin.POST("/system/settings/profile", backend_system.PostProfile)

	backstage_admin.GET("/system/settings/login", backend_system.Login)
	backstage_admin.POST("/system/settings/login", backend_system.PostLogin)
	backstage_admin.GET("/system/settings/login/logs", backend_system.LoginLogs)
	backstage_admin.GET("/system/settings/login/logs/list", backend_system.LoginLogsList)

	backstage_admin.GET("/system/database", backend_system.Database)
	backstage_admin.GET("/system/database/index", backend_system.Database)
	backstage_admin.GET("/system/database/list", backend_system.DatabaseList)
	backstage_admin.GET("/system/database/update", backend_system.DatabaseUpdate)
	backstage_admin.GET("/system/database/cleans", backend_system.DatabaseClean)
	backstage_admin.POST("/system/database/clean", backend_system.PostDatabaseClean)
	backstage_admin.POST("/system/database/delete", backend_system.PostDatabaseDelete)

	backstage_admin.GET("/system/database/clean_setting", backend_system.DatabaseCleanSetting)
	backstage_admin.POST("/system/database/clean_setting", backend_system.PostDatabaseCleanSetting)
	backstage_admin.GET("/system/db", backend_system.Db)
	backstage_admin.GET("/system/db/list", backend_system.DbNodeList)
	backstage_admin.GET("/system/db/add", backend_system.DbNodeAdd)
	backstage_admin.POST("/system/db/add", backend_system.PostDbNodeAdd)
	backstage_admin.GET("/system/db/details", backend_system.DbNodeDetails)
	backstage_admin.GET("/system/db/clean", backend_system.DbNodeClean)
	backstage_admin.GET("/system/db/logs", backend_system.DbNodeLogs)
	backstage_admin.GET("/system/db/update", backend_system.DbNodeUpdate)

	backstage_admin.GET("/tg", backend_tg.Home)
	backstage_admin.GET("/tg/add", backend_tg.Add)
	backstage_admin.POST("/tg/add", backend_tg.PostAdd)
	backstage_admin.GET("/tg/list", backend_tg.List)
	backstage_admin.POST("/tg/delete", backend_tg.PostSoftDelete)
	backstage_admin.GET("/tg/details", backend_tg.Details)
	backstage_admin.POST("/tg/trigger_status", backend_tg.TgbotTriggerStatus)
	backstage_admin.POST("/tg/trigger_listen_enable", backend_tg.TgbotTriggerListenEnable)

	backstage_admin.GET("/tg/log", backend_tg.Log)
	backstage_admin.GET("/tg/log/list", backend_tg.LogList)
	backstage_admin.POST("/tg/log/delete", backend_tg.LogDelete)

	backstage_admin.GET("/tg/banword", backend_tg.Banword)
	backstage_admin.GET("/tg/banword/list", backend_tg.BanwordList)
	backstage_admin.POST("/tg/banword/delete", backend_tg.BanwordDelete)

}

func initRuoteInstall(r *gin.Engine) {
	installGroup := r.Group("/install")
	installGroup.Use(middleware.CheckInstalledAfter())
	installGroup.GET("/index", install.HomePage)
	installGroup.POST("/step1", install.PostInstallStep1)
	installGroup.POST("/dbtest", install.MyDbtest)
}

func initRuoteFrontend(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.Use(middleware.CheckInstalled()).GET("/", home.Index)
}

func initRuote(r *gin.Engine) {
	// static files from embedded filesystem subdir "static"
	staticFS, err := fs.Sub(embed.Static, "static")
	if err != nil {
		panic(fmt.Sprintf("initRuote:%v", err))
	}
	// 设置静态文件缓存为一周
	staticHandler := http.StripPrefix("/static", http.FileServer(http.FS(staticFS)))
	r.GET("/static/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=604800") // 604800 秒 = 7 天
		// c.Header("Permissions-Policy", "unload=*")
		staticHandler.ServeHTTP(c.Writer, c.Request)
	})

	initRuoteInstall(r)
	initRuoteAdmin(r)
	initRuoteFrontend(r)
}

func Run() {
	if conf.App.RunMode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化清理任务
	op.InitCleanTask()

	r := gin.New()

	// 初始化 session 存储
	store := cookie.NewStore([]byte(conf.Security.SecretKey))
	// 设置 cookie 选项
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(conf.Session.MaxLifeTime),
		HttpOnly: true,
		Secure:   conf.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions(conf.Session.CookieName, store))

	// 启用压缩
	if conf.Web.EnableGzip {
		r.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)

	initTmplFunc(r)
	initRuote(r)
	r.Run(fmt.Sprintf(":%d", conf.Web.HTTPPort))
}
