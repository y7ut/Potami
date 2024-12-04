package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/y7ut/potami/pkg/json"
)

const (
	TableBoxWidth = 30
)

var (
	potamiPath  string
	contextPath string
	configPath  string
)

type PotamiServiceContext struct {
	Name        string `mapstructure:"name" json:"name"`
	Description string `mapstructure:"description" json:"description"`
	Endpoint    string `mapstructure:"endpoint" json:"endpoint"`
	Used        bool   `mapstructure:"used" json:"used"`
}

type PotamiConfig struct {
	CurrentContext string `mapstructure:"context" json:"context"`
}

var ContextCommand = &cobra.Command{
	Use:   "context",
	Short: "Potami Context",
}

var ContextListCommand = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Context",
	Run: func(c *cobra.Command, args []string) {
		contexts := getContexts()

		contextSlice := make([]PotamiServiceContext, 0, len(contexts))
		for _, k := range contexts {
			contextSlice = append(contextSlice, k)
		}
		slices.SortFunc(contextSlice, func(a, b PotamiServiceContext) int {
			return strings.Compare(a.Name, b.Name)
		})

		table := NewTable([]string{"NAME", "DESCRIPTION", "ENDPOINT"})
		for _, v := range contextSlice {
			if v.Used {
				v.Name = fmt.Sprintf(" * %s", v.Name)
			}
			table.AddRow([]string{v.Name, v.Description, v.Endpoint})
		}
		table.Render()
	},
}

var ContextSetCommand = &cobra.Command{
	Use:   "set",
	Short: "Set Context",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			c.Println("please set context name")
			return
		}
		contextToSet := args[0]
		contextUsed, err := setCurrentContextByName(contextToSet)
		if err != nil {
			c.Println(err)
			return
		}
		c.Printf("current context is %s\nendpoint is %s\n", contextUsed.Name, contextUsed.Endpoint)
	},
}

var ContextRemoveCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove Context",
	Run: func(c *cobra.Command, args []string) {
		if len(args) == 0 {
			c.Println("please set context name")
			return
		}
		contextToRemove := args[0]

		// check context exists
		contexts := getContexts()
		if _, ok := contexts[contextToRemove]; !ok {
			c.Printf("context %s not found\n", contextToRemove)
			return
		}
		if len(contexts) == 1 {
			c.Println("cannot remove last context")
			return
		}

		if contexts[contextToRemove].Used {
			c.Println("cannot remove used context")
			return
		}

		if err := removeContext(contextToRemove); err != nil {
			c.Println(err)
			return
		}

		c.Printf("context %s removed\n", contextToRemove)
	},
}

var ContextAddCommand = &cobra.Command{
	Use:   "add",
	Short: "Add Context",
	Run: func(c *cobra.Command, args []string) {
		if len(args) != 3 {
			c.Println("please set context name, endpoint and description")
			return
		}
		name := args[0]
		endpoint := args[1]
		description := args[2]
		endpointUrl, err := url.Parse(endpoint)
		if endpointUrl.Scheme == "" {
			endpoint = "http://" + endpoint
		}
		if err != nil {
			c.Printf("invalid endpoint: %s\n", endpoint)
			return
		}
		contextToAdd := PotamiServiceContext{
			Name:        name,
			Description: description,
			Endpoint:    endpoint,
		}
		if err := addContext(contextToAdd); err != nil {
			c.Println(err)
			return
		}
		if c.Flag("used").Changed {
			if _, err := setCurrentContextByName(name); err != nil {
				c.Println(err)
				return
			}
			c.Printf("add context %s and set as used\nendpoint is %s\n", name, endpoint)
		} else {
			c.Printf("add context %s\nendpoint is %s\n", name, endpoint)
		}
	},
}

func init() {
	ContextCommand.AddCommand(ContextListCommand)
	ContextCommand.AddCommand(ContextSetCommand)
	ContextCommand.AddCommand(ContextRemoveCommand)
	ContextAddCommand.Flags().BoolP("used", "u", false, "set as used")
	ContextCommand.AddCommand(ContextAddCommand)

	RootCmd.AddCommand(ContextCommand)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get user dir: %v", err)
	}

	potamiPath = fmt.Sprintf("%s/.potami", userHomeDir)
	if _, err := os.Stat(potamiPath); os.IsNotExist(err) {
		err := os.Mkdir(potamiPath, os.ModePerm)
		if err != nil {
			log.Fatalf("failed to create %s: %v", potamiPath, err)
		}
	}

	configPath = fmt.Sprintf("%s/config", potamiPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := os.Mkdir(configPath, os.ModePerm)
		if err != nil {
			log.Fatalf("failed to create %s: %v", configPath, err)
		}
	}

	contextPath = fmt.Sprintf("%s/context", configPath)
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		err := os.Mkdir(contextPath, os.ModePerm)
		if err != nil {
			log.Fatalf("failed to create %s: %v", contextPath, err)
		}
	}
}

// generateDefaultConfigIfNotExist 如果不存在生成默认配置文件
func generateDefaultConfigIfNotExist() {
	configFile := fmt.Sprintf("%s/config.json", configPath)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		config := PotamiConfig{
			CurrentContext: "local",
		}
		configJson, err := json.Marshal(config)
		if err != nil {
			log.Fatalf("failed to marshal config: %v", err)
		}

		err = os.WriteFile(configFile, configJson, 0644)
		if err != nil {
			log.Fatalf("failed to write config: %v", err)
		}
	}
}

// addContext 添加新的上下文
func addContext(context PotamiServiceContext) error {
	contextFile := fmt.Sprintf("%s/%s.json", contextPath, context.Name)
	if _, err := os.Stat(contextFile); os.IsNotExist(err) {
		// create context file
		defaultContextJson, err := json.Marshal(context)
		if err != nil {
			log.Fatalf("failed to marshal default context: %v", err)
		}

		err = os.WriteFile(contextFile, defaultContextJson, 0644)
		if err != nil {
			log.Fatalf("failed to write default context: %v", err)
		}

		return nil
	}

	return fmt.Errorf("context has been added")
}

// removeContext 删除上下文
func removeContext(name string) error {
	contextFile := fmt.Sprintf("%s/%s.json", contextPath, name)
	if _, err := os.Stat(contextFile); os.IsNotExist(err) {
		return fmt.Errorf("context %s not found", name)
	}
	err := os.Remove(contextFile)
	if err != nil {
		return err
	}
	return nil
}

// getDefaultContextName 获取默认上下文的名字
func getDefaultContextName() string {
	configFile := fmt.Sprintf("%s/config.json", configPath)
	generateDefaultConfigIfNotExist()
	config, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	var configObj PotamiConfig
	err = json.Unmarshal(config, &configObj)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}
	return configObj.CurrentContext
}

// getCurrentContext 获取当前使用的上下文
func getCurrentContext() (*PotamiServiceContext, error) {
	contexts := getContexts()
	for _, v := range contexts {
		if v.Used {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no current context found")
}

// getContexts 获取所有的上下文
func getContexts() map[string]PotamiServiceContext {
	contexts := make(map[string]PotamiServiceContext)

	err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == ".json" {
			viper := viper.New()
			viper.SetConfigType("json")
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("error reading config file, %s", err)
			}
			var context PotamiServiceContext
			err = viper.Unmarshal(&context)
			if err != nil {
				return fmt.Errorf("error unmarshal config file, %s", err)
			}
			if context.Name == getDefaultContextName() {
				context.Used = true
			}
			contexts[context.Name] = context
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed to read context path: %v", err)
	}
	if len(contexts) == 0 {
		// add default
		defaultContext := PotamiServiceContext{
			Name:        "local",
			Description: "potami default context, http://127.0.0.1:6180",
			Endpoint:    "http://127.0.0.1:6180",
			Used:        true,
		}

		if err := addContext(defaultContext); err != nil {
			log.Fatalf("failed to add default context: %v", err)
		}
		contexts["local"] = defaultContext
		setCurrentContextByName("local")
	}
	return contexts
}

// SetCurrentContextByName 通过名称设置使用当前的上下文
func setCurrentContextByName(contextToSet string) (*PotamiServiceContext, error) {
	contexts := getContexts()
	var contextUsed *PotamiServiceContext
	for k, v := range contexts {
		if k == contextToSet {
			contextUsed = &v
			setDefaultContextName(contextToSet)
			break
		}
	}
	if contextUsed == nil {
		return nil, fmt.Errorf("context %s not found", contextToSet)
	}
	return contextUsed, nil
}

// setDefaultContextName 通过名称设置使用当前的默认上下文
func setDefaultContextName(name string) {
	configFile := fmt.Sprintf("%s/config.json", configPath)

	generateDefaultConfigIfNotExist()
	config, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var configObj PotamiConfig
	err = json.Unmarshal(config, &configObj)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	configObj.CurrentContext = name
	config, err = json.Marshal(configObj)
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}
	err = os.WriteFile(configFile, config, 0644)
	if err != nil {
		log.Fatalf("failed to write config: %v", err)
	}
}
