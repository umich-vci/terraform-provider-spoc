resource "spoc_client" "example_client" {
  name        = "exampleclient"
  server_name = "backupserver"
  domain      = "backupdomain"
  password    = "backupclientpassword"
}
