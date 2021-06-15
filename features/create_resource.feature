Feature: create a new resource
  As a unauthenticated user
  I want to be able to create a new resource in a database through an API endpoint
  So I can have increase my available resources to answer my needs and feed my strategies

  Scenario: doing a valid query to create a resource
    When I send "POST" request to "/resource"
    Then the response code should be 200
