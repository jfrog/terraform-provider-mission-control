resource "missioncontrol_access_federation_star" "my-star" {
  id = "JPD-1"
  entities = ["USERS", "GROUPS", "PERMISSIONS"]
  targets = [
    {
      id = "JPD-2"
      url = "http://myartifactory-2.jfrog.io/access"
      permission_filters = {
        include_patterns = ["some-regex"]
      }
    },
  ]
}