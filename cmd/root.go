package cmd

import (
	"errors"
	"fmt"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/nesv/go-dynect/dynect"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var username string
var password string
var account string
var cfgFile string
var logLevel string
var Client *dynect.Client
var requiredFlags = []string{
	"username",
	"password",
}

func setLogLevel() error {
	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		log.WithFields(log.Fields{
			"event": "INVALID_LOG_LEVEL",
			"err":   err,
		}).Error("Given log level was invalid")
		return err
	}
	log.SetLevel(level)
	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dyn-cli-utils",
	Short: "CLI utils for the Dynect API",
	Long:  `Captures common tasks associated with configuration of the Dynect DNS API`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := setLogLevel()
		if err != nil {
			return err
		}

		var unsetFlags []string
		for _, flag := range requiredFlags {
			if !viper.IsSet(flag) || viper.GetString(flag) == "" {
				unsetFlags = append(unsetFlags, flag)
			}
		}
		if len(unsetFlags) > 0 {
			log.WithFields(log.Fields{
				"event": "MISSING_REQUIRED_FLAGS",
				"flags": unsetFlags,
			}).Error(fmt.Sprintf("'%s' flags are required", strings.Join(unsetFlags, ", ")))
			return errors.New("Missing required flags")
		}

		Client = dynect.NewClient(viper.GetString("account"))
		err = Client.Login(viper.GetString("username"), viper.GetString("password"))
		if err != nil {
			log.WithFields(log.Fields{
				"event": "DYNECT_API_LOGIN_FAILURE",
				"err":   err,
			}).Error("Dynect login failed")
			return err
		}
		log.WithFields(log.Fields{
			"event": "DYNECT_API_LOGIN_SUCCESS",
		}).Debug("Dynect login success")
		return nil
	},
	PersistentPostRunE: func(cmdb *cobra.Command, args []string) error {
		err := Client.Logout()
		if err != nil {
			log.WithFields(log.Fields{
				"event": "DYNECT_API_LOGOUT_FAILURE",
				"err":   err,
			}).Error("Dynect logout failed")
			return err
		}
		log.WithFields(log.Fields{
			"event": "DYNECT_API_LOGOUT_SUCCESS",
		}).Debug("Dynect logout successful")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"event": "ERROR_STARTING_CLI",
			"err":   err,
		}).Fatal("CLI failed to start")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dyn-cli.yaml)")
	RootCmd.PersistentFlags().StringVar(&username, "username", "", "the Dynect API username")
	RootCmd.PersistentFlags().StringVar(&password, "password", "", "the Dynect API password")
	RootCmd.PersistentFlags().StringVar(&account, "account", "financialtimes", "the Dynect API account")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "the log level to use")

	_ = viper.BindPFlag("username", RootCmd.PersistentFlags().Lookup("username"))
	_ = viper.BindPFlag("password", RootCmd.PersistentFlags().Lookup("password"))
	_ = viper.BindPFlag("account", RootCmd.PersistentFlags().Lookup("account"))
	_ = viper.BindPFlag("log-level", RootCmd.PersistentFlags().Lookup("log-level"))
}

func readEnvVariables() {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("Failed to determine $HOME directory")
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".dyn-cli")
	}
	readEnvVariables()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}
