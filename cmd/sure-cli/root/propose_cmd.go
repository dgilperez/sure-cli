package root

import (
	"time"

	"github.com/dgilperez/sure-cli/internal/api"
	"github.com/dgilperez/sure-cli/internal/output"
	"github.com/dgilperez/sure-cli/internal/rules"
	"github.com/spf13/cobra"
)

func newProposeCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "propose", Short: "Propose automations (rules)"}
	cmd.AddCommand(newProposeRulesCmd())
	return cmd
}

func newProposeRulesCmd() *cobra.Command {
	var months int

	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Propose categorization rules based on transaction patterns",
		Run: func(cmd *cobra.Command, args []string) {
			client := api.New()

			if months <= 0 {
				months = 3
			}
			end := time.Now().UTC()
			start := end.AddDate(0, -months, 0)
			txs, err := api.FetchTransactionsWindow(client, start, end, 500)
			if err != nil {
				output.Fail("request_failed", err.Error(), nil)
			}

			result := rules.ProposeRules(txs)
			_ = output.Print(format, output.Envelope{Data: result, Meta: &output.Meta{Schema: "docs/schemas/v1/propose_rules.schema.json", Status: 200}})
		},
	}
	cmd.Flags().IntVar(&months, "months", 3, "lookback months")
	return cmd
}
