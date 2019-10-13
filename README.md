# Building REST APIs with Go and TDD

This is a presentation from the Norfolk Gophers meetup where we talked about bulding a REST API within Go using Test-Driven-Development (TDD).

## What we are building

We are going to cover building an API that allows the following:

* Guests can signup with an email and password
* Can send the email and password to return a JSON Web Token (JWT)
* As a user, can create a new team
* As a user, can create a new team
* Can see all of your teams as a user

This will not be a complete application. There are many things to consider such as user roles and permissions, those are outside of the scope of this demo. Our goal is to get the basic concepts of building an API and more importantly, structuring the code for easy testing.

### Considerations

* Storing passwords securely
* Using envrionment variables, never hard code credentials like database or keys
* Preparing for deployment and testing locally (with Docker)

## Our First Test

When using TDD, especially if you are new, we need to make testing as easy as possible. We should also choose a test that is important and gives us the biggest advantage.

> Note: Exit early and often applies to testing as well, write the easiest test you can and _always_ test the happy path and sad path.
