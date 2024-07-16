resource "missioncontrol_access_federation_mesh" "my-mesh" {
  ids = ["JPD-1", "JPD-2"]
  entities = ["USERS", "GROUPS", "PERMISSIONS"]
}