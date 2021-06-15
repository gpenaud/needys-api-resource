Feature: delete an existing resource
  As a unauthenticated user
  I want to be able to delete an existing resource from a database through an API endpoint
  So I can remove current resources to conform with my current reality

  Scenario: doing a valid query to create a resource
    When I send "DELETE" request to "/product/{id}"
    Then the response code should be 200
