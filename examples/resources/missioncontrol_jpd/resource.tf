resource "missioncontrol_jpd" "my-jpd" {
  name = "MyJPD"
  url  = "https://myinstance.jfrog.io/"
  token  = "my-join-key"

  location = {
    city_name = "San Francisco"
    country_code = "US"
    latitude = 37.7749
    longitude = 122.4194
  }

  tags = [
    "prod",
    "dev",
  ]
}