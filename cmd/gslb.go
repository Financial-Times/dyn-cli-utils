package cmd

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Financial-Times/dyn-cli-utils/internal"
	"github.com/Financial-Times/dyn-cli-utils/internal/dyn"
	"github.com/spf13/viper"
)

var (
	fqdn         string
	labelPattern string
	serveMode    string
	wait         bool
)

var baseCmd = &cobra.Command{
	Use:   "gslb",
	Short: "Commands relating to Global Server Load Balancing service (GSLB) services",
	Long:  `gslb requires a subcommand, e.g. ` + "`dyn-cli-utils gslb update-pools`.",
	RunE:  nil,
}

func setCommonFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&fqdn, "fqdn", "prometheus.in.ft.com",
		"The full DNS name to edit the GSLB region for. See https://help.dyn.com/get-baseCmd-regions-api/")

	_ = viper.BindPFlag("fqdn", cmd.Flags().Lookup("fqdn"))
	return cmd
}

var subCommands = []*cobra.Command{
	func() *cobra.Command {
		cmd := &cobra.Command{
			Use:   "update-pools",
			Short: "Update a GSLB service pools",
			Long: `Updates Global Server Load Balancing service pools across all regions. ` +
				`Only pools which match the given label-pattern regular expression are updated.`,
			RunE: updatePools,
		}
		cmd.Flags().StringVar(&labelPattern, "label-pattern", "eu-[0-9]+",
			"A fully anchored regex (https://github.com/google/re2/wiki/Syntax) to match pool entry labels against. Matches the pool entry labels for each region.")
		cmd.Flags().StringVar(&serveMode, "serve-mode", "obey",
			"The serve_mode to set all pools in the given region to (always|obey|remove|no). See https://help.dyn.com/get-cmd-regions-api/")
		cmd.Flags().BoolVar(&wait, "wait", true, "Whether to wait for the GSLB TTL to return")

		_ = viper.BindPFlag("label-pattern", cmd.Flags().Lookup("label-pattern"))
		_ = viper.BindPFlag("serve-mode", cmd.Flags().Lookup("serve-mode"))
		_ = viper.BindPFlag("wait", cmd.Flags().Lookup("wait"))

		setCommonFlags(cmd)

		return cmd
	}(),
}

func init() {
	baseCmd.AddCommand(subCommands...)
	RootCmd.AddCommand(baseCmd)
}

func updatePools(cmd *cobra.Command, args []string) error {
	wait := viper.GetBool("wait")

	labelRegexp, err := regexp.Compile(fmt.Sprintf("^%s$", viper.GetString("label-pattern")))
	if err != nil {
		log.WithFields(log.Fields{
			"event": "GSLB_INVALID_LABEL_MATCH_REGEXP",
			"err":   err,
		}).Error("The provided 'label-match' regular expression was invalid")
		return err
	}
	config := dyn.GslbConfig{
		LabelPattern: labelRegexp,
		ServeMode:    viper.GetString("serve-mode"),
		Fqdn:         viper.GetString("fqdn"),
	}

	service := dyn.NewDynectService(Client)
	err = service.UpdateGslbRegion(config)
	if err != nil {
		log.WithFields(log.Fields{
			"event":  "UPDATE_GSLB_REGION_FAILURE",
			"err":    err,
			"config": config,
		}).Error("Failed to update GSLB region")
		return err
	}
	log.WithFields(log.Fields{
		"event":  "RECORD_UPDATE_SUCCESSFUL",
		"config": config,
	}).Info("Successfully updated GSLB region records")

	ttl, err := service.GetGslbTTL(config)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "DYNECT_GSLB_TTL_NOT_ESTABLISHED",
			"err":   err,
		}).Error("Could not get GSLB TTL for service")
		return err
	}

	if wait {
		log.WithFields(log.Fields{
			"ttl": ttl,
		}).Info("Waiting for DNS changes to propagate")
		terminal.WaitWithSpinner(ttl)
	}
	log.Info("🚀\tSuccess!")
	return nil
}
