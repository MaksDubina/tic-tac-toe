# Project Backend 05 — Go_Bootcamp

**Summary:** In this project, you will learn how to work with JWT-based authorization and extend the functionality of your Go web application using the net/http package.

💡 [Click here](https://new.oprosso.net/p/4cb31ec3f47a4596bc758ea1861fb624) to give us feedback on this project. It’s anonymous and will help our team improve the learning experience. We recommend completing the survey right after finishing the project.

## Table of Contents

  - [Chapter I](#chapter-i)
    - [Instructions](#instructions)
  - [Chapter II](#chapter-ii)
    - [General Information](#general-information)
  - [Chapter III](#chapter-iii)
    - [Project: Tic-Tac-Toe](#project-tic-tac-toe)
    - [Task 1: Replace Basic Authorization with JWT](#task-1-replace-basic-authorization-with-jwt)
    - [Task 2: Game History Support](#task-2-game-history-support)
    - [Task 3: Leaderboard Support](#task-3-leaderboard-support)

## Chapter I

### Instructions

1. Throughout this course, you will often feel uncertain and experience a lack of information — that’s normal. Don’t forget that the repository and Google are always at your disposal. So are your peers and Rocket.Chat. Communicate. Search. Use common sense. Don’t be afraid to make mistakes.
2. Be mindful of your information sources. Double-check. Think. Analyze. Compare.
3. Read the assignment carefully. Then read it again.
4. Read the examples carefully as well. They may contain important details that aren’t explicitly stated in the instructions.
5. You may encounter inconsistencies — something new in the task or example might contradict what you’ve already seen. If that happens, try to figure it out. If you can’t, write your question down under “Open Questions” and try to resolve it during the process. Don’t leave open questions unanswered.
6. If the task seems unclear or unachievable — it only seems that way. Try breaking it down. Most likely, individual parts will become clearer.
7. You’ll encounter various types of tasks. Those marked with an asterisk (\*) are optional and more advanced. They’re not required, but completing them will give you extra experience and knowledge.
8. Don’t try to cheat the system or others. You’ll only be cheating yourself.
9. Have a question? Ask the person to your right. If that doesn’t help, ask the one to your left.
10. When receiving help, always make sure you understand the what, how, and why. Otherwise, the help won’t be meaningful.
11. Always push your changes to the develop branch only! The master branch will be ignored. Work in the src directory.
12. Your directory should contain only those files specified in the assignment.

## Chapter II

### General Information

**Token, Session Token, Refresh Token**

A **token** is a unique string of characters that replaces the user’s login and password to prevent leakage of confidential information. Tokens have a limited lifespan after which they become invalid.

A **session token** grants the user rights to perform actions available to them during a session. It is reusable and has a short expiration time.

A **refresh token** extends the validity period of a session token. It is single-use and has a long expiration time.

**Topics to study:**

- Web applications
- JWT authorization
- PostgreSQL
- net/http

## Chapter III

### Project: Tic-Tac-Toe

Use the server-side project from the previous week, **T04**.

### Task 1: Replace Basic Authorization with JWT

- Create a JwtRequest model containing login and password.
- Create a JwtResponse model containing type, accessToken, and refreshToken.
- Create a RefreshJwtRequest model containing refreshToken.
- Implement a JwtProvider structure with the following methods:
  - A method to generate an accessToken based on a User; the token must include the user’s UUID.
  - A method to generate a refreshToken based on a User; the token must include the user’s UUID.
  - A method to validate the accessToken.
  - A method to validate the refreshToken.
  - A method to extract the UUID from a token.
- Update the **authorization service**, which uses UserService and JwtProvider, to include:
  - A modified authorization method that now accepts a JwtRequest and returns a JwtResponse.
  - A method to refresh the accessToken, which accepts a refreshToken and returns a JwtResponse.
  - A method to refresh the refreshToken, which accepts a refreshToken and returns a JwtResponse.
- Update the **authorization controller**:
  - Add or modify endpoints for:
    - user authorization
    - accessToken refresh
    - refreshToken refresh
- Change the logic for determining the authorized user:
  - Extract the token from the Authorization header, which contains Bearer {accessToken}.
  - Use JwtProvider to validate the token.
  - Set the authorization using the sign method of the JWT extension for Request.
  - Remove Basic Authorization from Authentication.
  - Add Bearer Authorization to Authentication.
  - Use JwtProvider to validate the token.
  - If validation fails, respond with status code 401 and do not execute the request.
  - Allow unauthenticated access to the accessToken refresh endpoint.
  - Add an endpoint to retrieve user info based on the accessToken.

### Task 2: Game History Support

- Add a creation date field to the game model.
- Define a database query that retrieves all completed games by user UUID.
- A game is considered completed if it has one of the following states:
  - Victory of the player with UUID;
  - Draw.
- In the game service, add a method that retrieves all completed games by the user’s UUID.
- Add an endpoint to retrieve all completed games based on the accessToken, accessible only to authorized users.

### Task 3: Leaderboard Support

- Create a model for information about won games, containing the user's UUID and their win ratio.
- Define a database query that:
  - Calculates the ratio of the number of wins to losses and draws for each user.
  - Sorts the win ratios in descending order.
  - Selects the top N records, each containing the user's UUID and win ratio.
- In the game service, add a method to retrieve the top N players.
- Add an endpoint that accepts N (number of top players) and returns a list of the best players (UUID and login) with their win ratios.
- The endpoint for retrieving top players must be accessible only to authorized users.