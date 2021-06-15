Feature: fetch all existing resources
  As a unauthenticated user
  I want to be able to fetch every existing resource from a database through an API endpoint
  So I can remember of all available resources to feed my strategies and answer my needs

  Scenario: doing a valid query to create a resource
    When I send "GET" request to "/products"
    Then the response code should be 200
