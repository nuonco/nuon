output "info" {
  value = module.repo.all
}

output "name" {
  value = var.ecr_repo_name
}

// outputs below here are for testing all possible
// types and nesting for output parsing code
output "sample_scalar_boolean" {
  value = true
}

output "sample_scalar_number" {
  value = -42.42
}

output "sample_scalar_string" {
  value = "this is a sample scalar string"
}

output "sample_array_of_strings" {
  value = ["alpha", "bravo"]
}

output "sample_array_mixed" {
  value = ["first element", true, 42]
}

output "sample_shallow_object" {
  value = {
    key1 = true
    key2 = 42.42
    key3 = "value1"
  }
}

output "sample_array_of_objects" {
  value = [{"key1": "value1"}, {"key2": "value2"}]
}

output "sample_nested_objects" {
  value = {
    key1 = "value1"
    key2 = "value2"
    key3 = {
      key3B1 = "value3"
      key3B2 = "value4"
      key3B3 = {
        key3C1 = 42.42
      }
    }
  }
}

output "sample_nested_objects_and_arrays" {
  value = {
    settings = [{
      host = "test.example.com"
      port = 8080
    }, {
      host = "test2.example.com"
      port = 9090
      theme = {
        bg = "white"
        fg = "black"
      }
    }]
  }
}
