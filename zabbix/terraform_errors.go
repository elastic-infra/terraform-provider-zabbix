package zabbix

import "github.com/hashicorp/terraform-plugin-sdk/v2/diag"

type TerraformErrors struct {
	errors []error
}

func (e *TerraformErrors) addError(err error) {
	e.errors = append(e.errors, err)
}

func (e *TerraformErrors) addFromTerraformErrors(errors TerraformErrors) {
	e.errors = append(e.errors, errors.errors...)
}

func (e *TerraformErrors) getDiagnostics() diag.Diagnostics {
	var diags diag.Diagnostics
	for _, err := range e.errors {
		if err != nil {
			diags = append(diags, diag.Diagnostic{Severity: diag.Error, Summary: err.Error()})
		}
	}
	return diags
}
