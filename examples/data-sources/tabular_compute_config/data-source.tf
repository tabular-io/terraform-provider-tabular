data "tabular_warehouse" "test" {
  name = "example-warehouse"
}

data "tabular_compute_config" "test" {
	warehouse_id = tabular_warehouse.test.id
}

output "spark_config" {
  value = data.tabular_compute_config.test.spark_config
}
