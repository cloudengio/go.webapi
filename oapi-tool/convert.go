package main

import (
	"context"
	"encoding/json"
	"os"

	"cloudeng.io/webapi/openapi"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

type ConvertFlags struct {
	Output        string `subcmd:"output,,'filename to save the formatted file as, if not set, the contents will be written to stdout'"`
	ConvertToYAML bool   `subcmd:"yaml,false,the output will be in YAML format regardless of the input format"`
}

func convertCmd(_ context.Context, values any, args []string) error {
	fv := values.(*ConvertFlags)
	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var doc2 openapi2.T
	if err := json.Unmarshal(data, &doc2); err != nil {
		return err
	}
	doc3, err := openapi2conv.ToV3(&doc2)
	if err != nil {
		return err
	}
	data, err = openapi.FormatV3(doc3, fv.ConvertToYAML)
	if err != nil {
		return err
	}
	if len(fv.Output) == 0 {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(fv.Output, data, 0660)
}
