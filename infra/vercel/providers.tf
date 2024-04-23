provider "vercel" {
  team = "nuonco"

  # Or omit this for the api_token to be read
  # from the VERCEL_API_TOKEN environment variable
  api_token = var.vercel_api_token
}
