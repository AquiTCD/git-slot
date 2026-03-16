package cmd

import (
	"fmt"
	"io"

	"github.com/AquiTCD/git-slot/internal/slot"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [slot]",
	Short: "Show detailed slot status",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useJSON, _ := cmd.Flags().GetBool("json")

		a, err := bootstrap()
		if err != nil {
			return err
		}

		slotName := ""
		if len(args) == 1 {
			slotName = args[0]
		}
		return runStatus(a.mgr, slotName, cmd.OutOrStdout(), useJSON)
	},
}

func init() {
	statusCmd.Flags().Bool("json", false, "Output in JSON format")

	rootCmd.AddCommand(statusCmd)
}

func runStatus(mgr slot.SlotManager, slotName string, out io.Writer, useJSON bool) error {
	if slotName == "" {
		statuses, err := mgr.StatusAll()
		if err != nil {
			return err
		}
		if useJSON {
			items := make([]jsonSlotStatus, len(statuses))
			for i, s := range statuses {
				items[i] = statusToJSON(s)
			}
			return writeJSON(out, items)
		}
		for i, st := range statuses {
			if i > 0 {
				_, _ = fmt.Fprintln(out)
			}
			printSlotStatus(st, out)
		}
		return nil
	}

	st, err := mgr.Status(slotName)
	if err != nil {
		return err
	}
	if useJSON {
		return writeJSON(out, statusToJSON(*st))
	}
	printSlotStatus(*st, out)
	return nil
}

func printSlotStatus(st slot.SlotStatus, out io.Writer) {
	_, _ = fmt.Fprintf(out, "Slot:    %s\n", st.Name)
	_, _ = fmt.Fprintf(out, "State:   %s\n", st.DisplayState())
	if st.State == slot.SlotActive {
		_, _ = fmt.Fprintf(out, "Branch:  %s\n", st.Branch)
		_, _ = fmt.Fprintf(out, "HEAD:    %s", st.HeadHash)
		if st.CommitSubject != "" {
			_, _ = fmt.Fprintf(out, " (%s)", st.CommitSubject)
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintf(out, "Path:    %s\n", st.Path)
		if len(st.Changes) > 0 {
			_, _ = fmt.Fprintln(out, "Changes:")
			for _, c := range st.Changes {
				_, _ = fmt.Fprintf(out, "  %s\n", c)
			}
		}
	}
}
