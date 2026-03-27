package tui

// FieldDef defines a configuration field that should be prompted for.
type FieldDef struct {
	Name     string
	Label    string
	Required bool
	Secret   bool   // mask input
	Default  string // default value
}

// ProviderFields returns the required/important fields for a provider account.
func ProviderFields(provider string) []FieldDef {
	common := []FieldDef{
		{Name: "name", Label: "Account name", Required: true},
	}

	switch provider {
	case "kubernetes":
		return append(common,
			FieldDef{Name: "context", Label: "Kubernetes context"},
			FieldDef{Name: "kubeconfigFile", Label: "Kubeconfig path", Default: "~/.kube/config"},
			FieldDef{Name: "namespaces", Label: "Namespaces (comma-separated, blank for all)"},
		)
	case "aws":
		return append(common,
			FieldDef{Name: "accountId", Label: "AWS Account ID", Required: true},
			FieldDef{Name: "assumeRole", Label: "IAM role to assume"},
			FieldDef{Name: "regions", Label: "Regions (comma-separated)", Default: "us-west-2"},
			FieldDef{Name: "defaultKeyPair", Label: "Default EC2 key pair"},
		)
	case "gcp":
		return append(common,
			FieldDef{Name: "project", Label: "GCP Project ID", Required: true},
			FieldDef{Name: "jsonPath", Label: "Service account JSON path", Secret: true},
			FieldDef{Name: "regions", Label: "Regions (comma-separated)"},
		)
	case "azure":
		return append(common,
			FieldDef{Name: "clientId", Label: "Client ID", Required: true},
			FieldDef{Name: "appKey", Label: "Client secret", Required: true, Secret: true},
			FieldDef{Name: "tenantId", Label: "Tenant ID", Required: true},
			FieldDef{Name: "subscriptionId", Label: "Subscription ID", Required: true},
			FieldDef{Name: "defaultResourceGroup", Label: "Default resource group", Required: true},
			FieldDef{Name: "defaultKeyVault", Label: "Default key vault", Required: true},
		)
	case "cloudfoundry":
		return append(common,
			FieldDef{Name: "apiHost", Label: "CF API host", Required: true},
			FieldDef{Name: "user", Label: "Username", Required: true},
			FieldDef{Name: "password", Label: "Password", Required: true, Secret: true},
		)
	case "oracle":
		return append(common,
			FieldDef{Name: "compartmentId", Label: "Compartment ID"},
			FieldDef{Name: "userId", Label: "User ID"},
			FieldDef{Name: "fingerprint", Label: "Key fingerprint"},
			FieldDef{Name: "tenancyId", Label: "Tenancy ID"},
			FieldDef{Name: "region", Label: "Region"},
			FieldDef{Name: "sshPrivateKeyFilePath", Label: "SSH private key path", Secret: true},
		)
	}
	return common
}

// StorageFields returns the required/important fields for a storage backend.
func StorageFields(storage string) []FieldDef {
	switch storage {
	case "s3":
		return []FieldDef{
			{Name: "bucket", Label: "S3 bucket name", Required: true},
			{Name: "region", Label: "AWS region", Default: "us-west-2"},
			{Name: "rootFolder", Label: "Root folder", Default: "front50"},
			{Name: "accessKeyId", Label: "Access key ID"},
			{Name: "secretAccessKey", Label: "Secret access key", Secret: true},
			{Name: "endpoint", Label: "Custom S3 endpoint"},
		}
	case "gcs":
		return []FieldDef{
			{Name: "bucket", Label: "GCS bucket name", Required: true},
			{Name: "project", Label: "GCP project ID", Required: true},
			{Name: "jsonPath", Label: "Service account JSON path", Required: true, Secret: true},
			{Name: "rootFolder", Label: "Root folder", Default: "front50"},
			{Name: "bucketLocation", Label: "Bucket location"},
		}
	case "azs":
		return []FieldDef{
			{Name: "storageAccountName", Label: "Storage account name", Required: true},
			{Name: "storageAccountKey", Label: "Storage account key", Required: true, Secret: true},
			{Name: "storageContainerName", Label: "Container name", Required: true},
		}
	case "oracle":
		return []FieldDef{
			{Name: "namespace", Label: "Object Storage namespace", Required: true},
			{Name: "compartmentId", Label: "Compartment ID", Required: true},
			{Name: "userId", Label: "User ID", Required: true},
			{Name: "fingerprint", Label: "Key fingerprint", Required: true},
			{Name: "sshPrivateKeyFilePath", Label: "SSH private key path", Required: true, Secret: true},
			{Name: "tenancyId", Label: "Tenancy ID", Required: true},
			{Name: "bucketName", Label: "Bucket name"},
			{Name: "region", Label: "Region"},
		}
	case "redis":
		return []FieldDef{
			{Name: "host", Label: "Redis host", Default: "localhost"},
			{Name: "port", Label: "Redis port", Default: "6379"},
		}
	}
	return nil
}

// CIFields returns the required/important fields for a CI master/account.
func CIFields(ciType string) []FieldDef {
	common := []FieldDef{
		{Name: "name", Label: "Master name", Required: true},
	}

	switch ciType {
	case "jenkins":
		return append(common,
			FieldDef{Name: "address", Label: "Jenkins URL", Required: true},
			FieldDef{Name: "username", Label: "Username"},
			FieldDef{Name: "password", Label: "Password", Secret: true},
			FieldDef{Name: "csrf", Label: "Enable CSRF (true/false)", Default: "true"},
		)
	case "concourse":
		return append(common,
			FieldDef{Name: "url", Label: "Concourse URL", Required: true},
			FieldDef{Name: "username", Label: "Username"},
			FieldDef{Name: "password", Label: "Password", Secret: true},
		)
	case "travis":
		return append(common,
			FieldDef{Name: "address", Label: "Travis API address"},
			FieldDef{Name: "baseUrl", Label: "Travis base URL"},
			FieldDef{Name: "githubToken", Label: "GitHub token", Secret: true},
		)
	case "wercker":
		return append(common,
			FieldDef{Name: "address", Label: "Wercker API URL", Required: true},
			FieldDef{Name: "user", Label: "Username", Required: true},
			FieldDef{Name: "token", Label: "API token", Required: true, Secret: true},
		)
	case "gcb":
		return append(common,
			FieldDef{Name: "project", Label: "GCP project ID", Required: true},
			FieldDef{Name: "subscriptionName", Label: "Pub/Sub subscription", Required: true},
			FieldDef{Name: "jsonKey", Label: "Service account JSON path", Required: true, Secret: true},
		)
	case "codebuild":
		return append(common,
			FieldDef{Name: "accountId", Label: "AWS account ID", Required: true},
			FieldDef{Name: "region", Label: "AWS region", Required: true},
			FieldDef{Name: "assumeRole", Label: "IAM role to assume"},
		)
	}
	return common
}
