data "tabular_warehouse" "warehouse" {
  name = "funhouse"
}

resource "tabular_database" "database" {
  warehouse_id = data.tabular_warehouse.warehouse.id
  name         = "other"
}