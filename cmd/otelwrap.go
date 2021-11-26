package main

import (
	"errors"
	"fmt"
	"github.com/QuangTung97/otelwrap"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := &cobra.Command{
		Use: "otelwrap",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing directory and interface list")
			}
			if args[0] != "." {
				return errors.New("only support '.' as directory")
			}

			out, err := cmd.Flags().GetString("out")
			if err != nil {
				return err
			}
			if out == "" {
				return errors.New("missing 'out' flag")
			}

			pkgName, err := cmd.Flags().GetString("pkg")
			if err != nil {
				return err
			}

			return otelwrap.RunCommand(otelwrap.CommandArgs{
				Dir:            args[0],
				SrcFileName:    os.Getenv("GOFILE"),
				InterfaceNames: args[1:],
				InAnother:      otelwrap.CheckInAnother(out),
				PkgName:        pkgName,
			}, out)
		},
	}
	cmd.Flags().String("out", "", "required, output file name")
	cmd.Flags().String("pkg", "", "package name if specified interface is in another package")

	err := cmd.Execute()
	if err != nil {
		fmt.Println("ERROR:", err)
	}
}
