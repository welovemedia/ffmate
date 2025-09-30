package cmd

import (
	"encoding/json"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/welovemedia/ffmate/v2/internal"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/yosev/debugo"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"

	_ "goyave.dev/goyave/v5/database/dialect/postgres"
	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the server",
	Run:   server,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().String("ffmpeg", "", "path to ffmpeg binary")
	serverCmd.Flags().Uint("port", 3000, "the port to listen to")
	if runtime.GOOS == "windows" {
		serverCmd.Flags().String("database", "%APPDATA%\\ffmate\\db.sql", "the path do the database")
	} else {
		serverCmd.Flags().String("database", "~/.ffmate/db.sqlite", "the path do the database")
	}
	serverCmd.Flags().Uint("max-concurrent-tasks", 3, "define maximum concurrent running tasks")
	serverCmd.Flags().Bool("tray", false, "start with tray menu (experimental)")
	serverCmd.Flags().Bool("send-telemetry", true, "enable sending anonymous telemetry data")
	serverCmd.Flags().Bool("no-ui", false, "do not open the ui in the browser")
	serverCmd.Flags().String("identifier", "", "a unique client identifier (default to hostname)")
	serverCmd.Flags().StringSlice("labels", []string{}, "a unique client identifier (default to hostname)")
	serverCmd.Flags().String("basic-auth", "", "enable basic auth in form of username:password")

	_ = viper.BindPFlag("ffmpeg", serverCmd.Flags().Lookup("ffmpeg"))
	_ = viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
	_ = viper.BindPFlag("database", serverCmd.Flags().Lookup("database"))
	_ = viper.BindPFlag("maxConcurrentTasks", serverCmd.Flags().Lookup("max-concurrent-tasks"))
	_ = viper.BindPFlag("tray", serverCmd.Flags().Lookup("tray"))
	_ = viper.BindPFlag("sendTelemetry", serverCmd.Flags().Lookup("send-telemetry"))
	_ = viper.BindPFlag("noUI", serverCmd.Flags().Lookup("no-ui"))
	_ = viper.BindPFlag("identifier", serverCmd.Flags().Lookup("identifier"))
	_ = viper.BindPFlag("labels", serverCmd.Flags().Lookup("labels"))
	_ = viper.BindPFlag("basicAuth", serverCmd.Flags().Lookup("basic-auth"))
}

func server(_ *cobra.Command, _ []string) {
	setupConfig()

	// init goyave with config
	internal.Init(goyave.Options{
		Config: setupGoyaveConfig(),
	})
}

func setupConfig() {
	isCluster := strings.HasPrefix(viper.GetString("database"), "postgresql://")

	// docker
	_, err := os.Stat("/.dockerenv")
	cfg.Set("ffmate.isDocker", err == nil)

	// client identifier
	client := viper.GetString("identifier")
	if client == "" {
		client, err = os.Hostname()
		if err != nil {
			panic(err)
		}
	}

	labels := viper.GetStringSlice("labels")
	for i := range labels {
		labels[i] = strings.TrimSpace(labels[i])
	}

	cfg.Set("ffmate.ffmpeg", viper.GetString("ffmpeg"))
	cfg.Set("ffmate.debug", viper.GetString("debug"))
	cfg.Set("ffmate.maxConcurrentTasks", viper.GetInt("maxConcurrentTasks"))
	cfg.Set("ffmate.database", viper.GetString("database"))
	cfg.Set("ffmate.isTray", viper.GetBool("tray"))
	cfg.Set("ffmate.isUI", !viper.GetBool("noUI"))
	cfg.Set("ffmate.isCluster", isCluster)
	cfg.Set("ffmate.isAuth", false)
	cfg.Set("ffmate.isFFmpeg", false)
	cfg.Set("ffmate.identifier", client)
	cfg.Set("ffmate.labels", labels)
	cfg.Set("ffmate.session", uuid.NewString())
	cfg.Set("ffmate.telemetry.send", viper.GetBool("sendTelemetry"))
	cfg.Set("ffmate.telemetry.url", "https://telemetry.ffmate.io")

	debugo.SetNamespace(viper.GetString("debug"))
}

// create a temporary json config and pass it to goyave
func setupGoyaveConfig() *config.Config {
	// replace possible ~ with user home folder
	if strings.HasPrefix(viper.GetString("database"), "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		viper.Set("database", strings.Replace(viper.GetString("database"), "~", home, 1))
	}

	c := map[string]any{
		"app": map[string]any{
			"name":    viper.GetString("app.name"),
			"version": viper.GetString("app.version"),
			"debug":   false,
		},
		"server": map[string]any{
			"host": "0.0.0.0",
			"port": int(viper.GetUint("port")),
		},
		"database": map[string]any{},
		"auth": map[string]any{
			"basic": map[string]any{
				"username": "",
				"password": "",
			},
		},
	}

	// handle basic auth
	if viper.GetString("basicAuth") != "" {
		credentials := strings.SplitN(viper.GetString("basicAuth"), ":", 2)
		if len(credentials) == 2 {
			c["auth"].(map[string]any)["basic"].(map[string]any)["username"] = credentials[0]
			c["auth"].(map[string]any)["basic"].(map[string]any)["password"] = credentials[1]
			cfg.Set("ffmate.isAuth", true)
		} else {
			debug.Log.Error("invalid basic-auth configuration")
			os.Exit(1)
		}
	}

	// configure database
	if strings.HasPrefix(viper.GetString("database"), "postgresql://") {
		// parse postgresl uri to url
		url, err := url.Parse(viper.GetString("database"))
		if err != nil {
			panic(err)
		}

		c["database"].(map[string]any)["connection"] = "postgres"
		c["database"].(map[string]any)["host"] = url.Host
		c["database"].(map[string]any)["name"] = strings.Trim(url.Path, "/")
		c["database"].(map[string]any)["port"] = 5432
		if url.Port() != "" {
			if port, err := strconv.Atoi(url.Port()); err == nil {
				c["database"].(map[string]any)["port"] = port
			} else {
				panic(err)
			}
		}
		c["database"].(map[string]any)["username"] = url.User.Username()
		if password, ok := url.User.Password(); ok {
			c["database"].(map[string]any)["password"] = password
		}
		c["database"].(map[string]any)["options"] = "?" + url.RawQuery
	} else {
		c["database"].(map[string]any)["connection"] = "sqlite3"
		c["database"].(map[string]any)["name"] = viper.GetString("database")
	}

	// marshal config to json
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	// init config from json string
	cfg, err := config.LoadJSON(string(b))
	if err != nil {
		panic(err)
	}

	return cfg
}
