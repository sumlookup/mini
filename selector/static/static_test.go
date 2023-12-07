package static_test

import (
	"os"
	"testing"

	selector "github.com/sumlookup/mini/selector/static"
)

func TestStatic(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "svc.cluster.local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[splitservice]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")

	s := selector.NewSelector()
	node, err := s.Select("core-role")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "role.core.svc.cluster.local:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}

}

func TestStaticLocalhost(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "svc.cluster.local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[splitservice]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")
	os.Setenv("ENV_STATIC_SELECTOR_PORT_NUMBER", "8080")

	s := selector.NewSelector()
	node, err := s.Select("127.0.0.1")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "127.0.0.1:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}

}

func TestStaticSimple(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[name]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")

	s := selector.NewSelector()
	node, err := s.Select("core-role")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "role.local:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}

func TestStaticSimpleCluster(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "default.svc.cluster.local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[name]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")

	s := selector.NewSelector()
	node, err := s.Select("core-role")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "role.default.svc.cluster.local:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}

func TestStaticWithMod(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "svc.cluster.local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[splitservice]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")
	os.Setenv("STATIC_SELECTOR_ENVMOD", "true")
	os.Setenv("ENV", "prod")

	s := selector.NewSelector()
	node, err := s.Select("core-role")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "role.core-prod.svc.cluster.local:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}

func TestStaticWithModMultiDash(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "svc.cluster.local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[splitservice]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")
	os.Setenv("STATIC_SELECTOR_ENVMOD", "true")
	os.Setenv("ENV", "dev")

	s := selector.NewSelector()
	node, err := s.Select("module-iso-simulator")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "iso-simulator.module-dev.svc.cluster.local:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}

func TestStaticWithDirect(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "svc.cluster.local")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "[direct]")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")
	os.Setenv("STATIC_SELECTOR_ENVMOD", "true")
	os.Setenv("ENV", "dev")

	s := selector.NewSelector()
	node, err := s.Select("module-iso-simulator")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "module-iso-simulator.svc.cluster.local:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}

func TestStaticNoEnv(t *testing.T) {

	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")

	s := selector.NewSelector()
	node, err := s.Select("core-role")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "core-role:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}

func TestStaticDocker(t *testing.T) {
	os.Setenv("STATIC_SELECTOR_DOMAIN_NAME", "")
	os.Setenv("STATIC_SELECTOR_SUFFIX", "")
	os.Setenv("DEFAULT_PORT_NUMBER", "8080")

	s := selector.NewSelector()
	node, err := s.Select("core-role")
	if err != nil {
		t.Error(err)
	}

	r, err := node()

	if err != nil {
		t.Error(err)
	}

	expect := "core-role:8080"
	if r.Address != expect {
		t.Errorf("invalid slector address, expected %s. got %s", expect, r.Address)
	}
}
