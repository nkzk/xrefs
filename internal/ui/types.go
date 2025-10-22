package ui

import (
	"fmt"
	"reflect"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func headersFromRow(r row) []string {
	v := reflect.ValueOf(r)
	t := v.Type()
	out := make([]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		out[i] = t.Field(i).Name
	}
	return out
}

func toStringRow(r row) []string {
	return []string{r.Namespace, r.Kind, r.ApiVersion, r.Name, r.Synced, r.SyncedReason, r.Ready, r.ReadyReason}
}

func toRow(s []string) (row, error) {
	var r row
	v := reflect.ValueOf(r)

	if len(s) != v.NumField() {
		return row{}, fmt.Errorf("row has %d fields but only %d is allowed", len(s), v.NumField())
	}

	rv := reflect.ValueOf(&r).Elem()
	for i := 0; i < rv.NumField(); i++ {
		v.Type().Field(i)
		f := rv.Field(i)
		sf := v.Field(i)

		if !f.CanSet() {
			return row{}, fmt.Errorf("field %q cannot be set (make it exported)", sf.Type().Name())
		}

		if f.Kind() != reflect.String {
			return row{}, fmt.Errorf("field %q is %s, need string", sf.Type().Name(), f.Kind())
		}

		f.SetString(s[i])
	}
	return r, nil
}

type resourceRef struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
	Name       string `json:"name" yaml:"name"`
}

type XR struct {
	Metadata struct {
		Namespace string `json:"namespace" yaml:"namespace"`
	} `json:"metadata" yaml:"metadata"`
	Spec struct {
		Crossplane struct {
			ResourceRefs []resourceRef `json:"resourceRefs" yaml:"resourceRefs"`
		} `json:"crossplane" yaml:"crossplane"`
	} `json:"spec" yaml:"spec"`
}
