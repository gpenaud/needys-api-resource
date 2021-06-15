Feature: fetch an existing resource
  As a unauthenticated user
  I want to be able to fetch an existing resource from a database through an API endpoint
  So I can have details about a resource

  Scenario: doing a valid query to create a resource
    When I send "GET" request to "/product/{id}"
    Then the response code should be 200
