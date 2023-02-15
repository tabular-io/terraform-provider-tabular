# Terraform Tabular Provider

## Using the provider

```
terraform {
  required_providers {
    tabular = {
      source = "tabular-io/tabular"
    }
  }
}

provider "tabular" {
  token_endpoint = "https://api.tabular.io/ws/v1/oauth/tokens"
  endpoint = "http://localhost:8080"
  organization_id = "73cade70-d578-4b84-8ba6-4cdebbdcf0f0"
}
```

