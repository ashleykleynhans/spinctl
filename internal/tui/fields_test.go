package tui

import "testing"

func TestProviderFieldsKubernetes(t *testing.T) {
	fields := ProviderFields("kubernetes")
	names := fieldNames(fields)
	for _, want := range []string{"name", "context", "kubeconfigFile", "namespaces"} {
		if !contains(names, want) {
			t.Errorf("kubernetes fields missing %q", want)
		}
	}
}

func TestProviderFieldsAws(t *testing.T) {
	fields := ProviderFields("aws")
	names := fieldNames(fields)
	for _, want := range []string{"name", "accountId", "regions"} {
		if !contains(names, want) {
			t.Errorf("aws fields missing %q", want)
		}
	}
	assertRequired(t, fields, "accountId")
}

func TestProviderFieldsGcp(t *testing.T) {
	fields := ProviderFields("gcp")
	names := fieldNames(fields)
	if !contains(names, "project") {
		t.Error("gcp fields missing project")
	}
	assertRequired(t, fields, "project")
}

func TestProviderFieldsAzure(t *testing.T) {
	fields := ProviderFields("azure")
	if len(fields) < 7 {
		t.Errorf("azure fields count = %d, want >= 7", len(fields))
	}
	for _, want := range []string{"clientId", "appKey", "tenantId", "subscriptionId"} {
		assertRequired(t, fields, want)
	}
}

func TestProviderFieldsCloudfoundry(t *testing.T) {
	fields := ProviderFields("cloudfoundry")
	for _, want := range []string{"apiHost", "user", "password"} {
		assertRequired(t, fields, want)
	}
}

func TestProviderFieldsOracle(t *testing.T) {
	fields := ProviderFields("oracle")
	if len(fields) < 2 {
		t.Errorf("oracle fields count = %d, want >= 2", len(fields))
	}
}

func TestProviderFieldsUnknown(t *testing.T) {
	fields := ProviderFields("unknown")
	if len(fields) != 1 {
		t.Errorf("unknown provider fields count = %d, want 1 (just name)", len(fields))
	}
	if fields[0].Name != "name" {
		t.Errorf("unknown provider field = %q, want name", fields[0].Name)
	}
}

func TestStorageFieldsS3(t *testing.T) {
	fields := StorageFields("s3")
	assertRequired(t, fields, "bucket")
}

func TestStorageFieldsGcs(t *testing.T) {
	fields := StorageFields("gcs")
	for _, want := range []string{"bucket", "project", "jsonPath"} {
		assertRequired(t, fields, want)
	}
}

func TestStorageFieldsAzs(t *testing.T) {
	fields := StorageFields("azs")
	required := 0
	for _, f := range fields {
		if f.Required {
			required++
		}
	}
	if required != 3 {
		t.Errorf("azs required count = %d, want 3", required)
	}
}

func TestStorageFieldsOracle(t *testing.T) {
	fields := StorageFields("oracle")
	required := 0
	for _, f := range fields {
		if f.Required {
			required++
		}
	}
	if required != 6 {
		t.Errorf("oracle storage required count = %d, want 6", required)
	}
}

func TestStorageFieldsRedis(t *testing.T) {
	fields := StorageFields("redis")
	m := fieldMap(fields)
	if m["host"].Default != "localhost" {
		t.Errorf("redis host default = %q, want localhost", m["host"].Default)
	}
	if m["port"].Default != "6379" {
		t.Errorf("redis port default = %q, want 6379", m["port"].Default)
	}
}

func TestStorageFieldsUnknown(t *testing.T) {
	fields := StorageFields("unknown")
	if fields != nil {
		t.Errorf("unknown storage fields = %v, want nil", fields)
	}
}

func TestCIFieldsJenkins(t *testing.T) {
	fields := CIFields("jenkins")
	assertRequired(t, fields, "address")
}

func TestCIFieldsWercker(t *testing.T) {
	fields := CIFields("wercker")
	for _, want := range []string{"address", "user", "token"} {
		assertRequired(t, fields, want)
	}
}

func TestCIFieldsUnknown(t *testing.T) {
	fields := CIFields("unknown")
	if len(fields) != 1 {
		t.Errorf("unknown CI fields count = %d, want 1", len(fields))
	}
	if fields[0].Name != "name" {
		t.Errorf("unknown CI field = %q, want name", fields[0].Name)
	}
}

// helpers

func fieldNames(fields []FieldDef) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = f.Name
	}
	return out
}

func fieldMap(fields []FieldDef) map[string]FieldDef {
	m := make(map[string]FieldDef, len(fields))
	for _, f := range fields {
		m[f.Name] = f
	}
	return m
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func assertRequired(t *testing.T, fields []FieldDef, name string) {
	t.Helper()
	for _, f := range fields {
		if f.Name == name {
			if !f.Required {
				t.Errorf("field %q should be required", name)
			}
			return
		}
	}
	t.Errorf("field %q not found", name)
}
