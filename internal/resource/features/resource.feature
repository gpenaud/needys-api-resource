Feature: manage resources
  As a unauthenticated user
  I want to be able to make CRUD operations in a database through an API endpoint
  So I can have manage my available resources to answer my needs and feed my strategies

  Scenario: doing a valid query to create a resource
    When I send "POST" request to "/resource"
    Then the response code should be 201

  Scenario: doing a valid query to delete a resource
    When I send "DELETE" request to "/resource/2"
    Then the response code should be 200

  Scenario: doing a valid query to fetch a list of resources
    When I send "GET" request to "/resources"
    Then the response code should be 200

  Scenario: doing a valid query to fetch a resource
    When I send "GET" request to "/resource/1"
    Then the response code should be 200

  Scenario: doing a valid query to update a resource
    When I send "PUT" request to "/resource/1"
    Then the response code should be 200
