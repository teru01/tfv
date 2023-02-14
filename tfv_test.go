package core

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTfv(t *testing.T) {
	basePath := "./testdata/case1"
	actualVariables, actualTfVars, err := generateVariables(true, basePath, basePath+"/variables.tf", basePath+"/main.tfvars")
	assert.NoError(t, err)

	expectedVariables, err := os.ReadFile(basePath + "/variables.sync.expected.tf")
	assert.NoError(t, err)
	expectedTfVars, err := os.ReadFile(basePath + "/main.tfvars.expected")
	assert.NoError(t, err)

	assert.Equal(t, string(expectedVariables), actualVariables, "variables file")
	assert.Equal(t, string(expectedTfVars), actualTfVars, "tfvars file")
}
