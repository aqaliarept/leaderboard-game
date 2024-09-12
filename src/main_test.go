package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func Test_application_container_deps_configuration(t *testing.T) {
	err := fx.ValidateApp(configureContainer())
	require.NoError(t, err)
}
