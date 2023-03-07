data "tabular_warehouse" "warehouse" {
  name = "example-warehouse"
}

output "warehouse_id" {
  value = data.tabular_warehouse.warehouse.id
}

output "warehouse_region" {
  value = data.tabular_warehouse.warehouse.region
}