<div id="top">

<!-- HEADER STYLE: CLASSIC -->
<div align="center">

# MYRIDERS-BE

<em>Accelerate Innovation, Power Seamless Ride Experiences</em>

<!-- BADGES -->
<img src="https://img.shields.io/github/license/nashirabbash/myRiders-BE?style=flat&logo=opensourceinitiative&logoColor=white&color=0080ff" alt="license">
<img src="https://img.shields.io/github/last-commit/nashirabbash/myRiders-BE?style=flat&logo=git&logoColor=white&color=0080ff" alt="last-commit">
<img src="https://img.shields.io/github/languages/top/nashirabbash/myRiders-BE?style=flat&color=0080ff" alt="repo-top-language">
<img src="https://img.shields.io/github/languages/count/nashirabbash/myRiders-BE?style=flat&color=0080ff" alt="repo-language-count">

<em>Built with the tools and technologies:</em>

<img src="https://img.shields.io/badge/JSON-000000.svg?style=flat&logo=JSON&logoColor=white" alt="JSON">
<img src="https://img.shields.io/badge/Markdown-000000.svg?style=flat&logo=Markdown&logoColor=white" alt="Markdown">
<img src="https://img.shields.io/badge/Go-00ADD8.svg?style=flat&logo=Go&logoColor=white" alt="Go">
<img src="https://img.shields.io/badge/Gin-008ECF.svg?style=flat&logo=Gin&logoColor=white" alt="Gin">
<img src="https://img.shields.io/badge/YAML-CB171E.svg?style=flat&logo=YAML&logoColor=white" alt="YAML">

</div>
<br>

---

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Testing](#testing)
- [Features](#features)
- [Project Structure](#project-structure)
  - [Project Index](#project-index)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgment](#acknowledgment)

---

## Overview

---

## Features

|     | Component         | Details                                                                                                                                                                        |
| :-- | :---------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ⚙️  | **Architecture**  | <ul><li>RESTful API server built with **Gin**</li><li>Layered structure: handlers, services, repositories</li><li>Database interactions via **sqlc** generated code</li></ul>  |
| 🔩  | **Code Quality**  | <ul><li>Consistent code style with **goimports** and **golangci-lint**</li><li>Strong typing and idiomatic Go patterns</li><li>Use of context for request management</li></ul> |
| 📄  | **Documentation** | <ul><li>Basic README with setup instructions</li><li>API endpoints documented via comments</li><li>Configurable via environment variables</li></ul>                            |
| 🔌  | **Integrations**  | <ul><li>Database: **PostgreSQL** (via connection pooling)</li><li>SQL migrations managed with **sqlc.yaml**</li><li>WebSocket support for real-time features</li></ul>         |
| 🧩  | **Modularity**    | <ul><li>Separate packages for handlers, services, repositories</li><li>Configurable via environment variables</li><li>Use of interfaces for dependency injection</li></ul>     |
| 🧪  | **Testing**       | <ul><li>Unit tests with **testify**</li><li>Mocked dependencies for isolation</li><li>Test coverage reports integrated</li></ul>                                               |
| ⚡️  | **Performance**   | <ul><li>Optimized database queries via **sqlc**</li><li>Connection pooling and context management</li><li>Minimal middleware for request handling</li></ul>                    |
| 🛡️  | **Security**      | <ul><li>Input validation and sanitization</li><li>JWT authentication (implied by typical backend patterns)</li><li>Secure environment variable management</li></ul>            |
| 📦  | **Dependencies**  | <ul><li>Managed with **go.mod** and **go.sum**</li><li>Includes libraries like **gin**, **protobuf**, **uuid**, **websocket**, **sonic**</li></ul>                             |

---

## Project Structure

```sh
└── myRiders-BE/
    ├── CLAUDE.md
    ├── SETUP.md
    ├── TASK.md
    ├── cmd
    │   └── server
    ├── go.mod
    ├── go.sum
    ├── internal
    │   ├── config
    │   ├── db
    │   ├── errors
    │   ├── handler
    │   ├── jobs
    │   ├── middleware
    │   ├── router
    │   ├── service
    │   └── websocket
    ├── pkg
    │   ├── jwt
    │   └── polyline
    ├── skills-lock.json
    └── sqlc.yaml
```

---

### Project Index

<details open>
	<summary><b><code>MYRIDERS-BE/</code></b></summary>
	<!-- __root__ Submodule -->
	<details>
		<summary><b>__root__</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ __root__</b></code>
			<table style='width: 100%; border-collapse: collapse;'>
			<thead>
				<tr style='background-color: #f8f9fa;'>
					<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
					<th style='text-align: left; padding: 8px;'>Summary</th>
				</tr>
			</thead>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/sqlc.yaml'>sqlc.yaml</a></b></td>
					<td style='padding: 8px;'>- Defines the configuration for generating Go database access code from PostgreSQL schemas and queries<br>- It streamlines database interactions by producing type-safe, JSON-serializable, and prepared query functions within the project’s architecture, facilitating efficient and maintainable data layer integration aligned with the overall system design.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/skills-lock.json'>skills-lock.json</a></b></td>
					<td style='padding: 8px;'>- Defines the projects core skill set and architectural principles, serving as a foundational reference for consistent API design, security, database schema, and development best practices<br>- It ensures alignment across various components and teams, facilitating scalable, secure, and maintainable system development within the broader codebase architecture.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/SETUP.md'>SETUP.md</a></b></td>
					<td style='padding: 8px;'>- Defines the core backend API endpoints and WebSocket communication channels for managing user authentication, vehicle data, and ride lifecycle events<br>- Facilitates real-time GPS data streaming during rides, enabling tracking, analysis, and leaderboard updates within the overall architecture<br>- Serves as the central interface connecting client interactions with database operations and real-time data processing.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/go.sum'>go.sum</a></b></td>
					<td style='padding: 8px;'>- The <code>go.sum</code> file ensures the integrity and consistency of the projects dependencies, specifically locking in the versions of testing libraries such as Ginkgo and Gomega<br>- These dependencies facilitate structured, behavior-driven testing within the codebase, supporting reliable validation of the applications components<br>- Overall, this file plays a crucial role in maintaining a stable testing environment that underpins the robustness of the entire project architecture.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/CLAUDE.md'>CLAUDE.md</a></b></td>
					<td style='padding: 8px;'>- The <code>CLAUDE.md</code> file serves as the comprehensive backend developer guide for the TrackRide project, which is built using Go and leverages PostgreSQL, WebSocket, and other modern technologies<br>- It outlines the core architecture and key components of the system, emphasizing its role in supporting real-time ride tracking and related functionalities<br>- This document provides essential context on the projects technical stack, API endpoints, and system design, ensuring developers understand how individual code files contribute to the overall backend infrastructure<br>- In essence, it acts as a blueprint for maintaining, extending, and integrating the backend services that power the TrackRide platform.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/go.mod'>go.mod</a></b></td>
					<td style='padding: 8px;'>- Defines project dependencies and module configuration for TrackRide, establishing core external libraries such as web frameworks, database drivers, and utility packages<br>- Serves as the foundational setup that enables seamless integration of components, ensuring consistent environment management and facilitating the development of a scalable, real-time ride tracking application.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/TASK.md'>TASK.md</a></b></td>
					<td style='padding: 8px;'>- The <code>TASK.md</code> file outlines the implementation plan for the TrackRide backend MVP, serving as a strategic guide for building a modular, scalable backend system<br>- It details the core objectives, including JWT authentication, real-time GPS streaming via WebSocket, social feed, weekly leaderboards, and push notifications<br>- The document emphasizes a domain-driven architecture with distinct modules such as handlers, services, middleware, database interactions, WebSocket management, and background jobs, facilitating independent development and testing<br>- Overall, this plan ensures a structured approach to delivering a robust backend that supports the key features of the TrackRide application.</td>
				</tr>
			</table>
		</blockquote>
	</details>
	<!-- internal Submodule -->
	<details>
		<summary><b>internal</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ internal</b></code>
			<!-- db Submodule -->
			<details>
				<summary><b>db</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.db</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/queries.go'>queries.go</a></b></td>
							<td style='padding: 8px;'>- Defines the database query interface and provides a factory function to instantiate query handlers using a PostgreSQL connection pool<br>- It facilitates seamless interaction with the database layer, enabling other components to perform data operations efficiently within the overall architecture<br>- This setup supports clean separation of concerns and promotes maintainability in the codebase.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/store.go'>store.go</a></b></td>
							<td style='padding: 8px;'>- Provides a foundational interface for database interactions within the project architecture, encapsulating connection pooling and query execution capabilities<br>- Serves as the primary access point for data operations, facilitating seamless integration with the applications data layer<br>- Acts as a bridge between the core application logic and the underlying PostgreSQL database, supporting future enhancements with generated query methods.</td>
						</tr>
					</table>
					<!-- sqlc Submodule -->
					<details>
						<summary><b>sqlc</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ internal.db.sqlc</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/social.sql.go'>social.sql.go</a></b></td>
									<td style='padding: 8px;'>- This code file defines database interactions related to ride comments within the applications social feature set<br>- Specifically, it provides functionality to create new comments associated with rides, enabling users to share feedback or discussions about specific rides<br>- As part of the overall architecture, this module facilitates user engagement and social interaction by managing comment data, integrating seamlessly with the ride and user data models to support a dynamic, community-driven experience.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/db.go'>db.go</a></b></td>
									<td style='padding: 8px;'>- Provides an abstraction layer for database interactions within the project, enabling execution of SQL commands, queries, and transactions<br>- Facilitates seamless communication with the PostgreSQL database through a unified interface, supporting both direct and transactional operations<br>- This component is essential for maintaining data consistency and integrity across the applications architecture.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/auth.sql.go'>auth.sql.go</a></b></td>
									<td style='padding: 8px;'>- Defines database operations for user management, including creating users, retrieving user details by email, ID, or username, and updating user profiles and push tokens<br>- Facilitates seamless integration between application logic and persistent storage, ensuring efficient user data handling within the overall architecture<br>- Supports core authentication and user profile functionalities essential for user-centric features.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/vehicles.sql.go'>vehicles.sql.go</a></b></td>
									<td style='padding: 8px;'>- Defines database operations for managing vehicle records, including creation, retrieval, updating, deactivation, and deletion, within the applications architecture<br>- Facilitates seamless interaction with the vehicles data, ensuring data integrity and supporting features like active vehicle listing and ride association checks, integral to the overall systems vehicle management functionality.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/rides.sql.go'>rides.sql.go</a></b></td>
									<td style='padding: 8px;'>- This code file provides database query functions related to managing ride records within the applications data layer<br>- Specifically, it includes functionality to mark a ride as abandoned, updating its status and timestamp details<br>- This supports the overall architecture by enabling consistent and reliable updates to ride lifecycle states, ensuring accurate tracking and management of user rides throughout the system.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/copyfrom.go'>copyfrom.go</a></b></td>
									<td style='padding: 8px;'>- Facilitates efficient bulk insertion of leaderboard entries into the database, enabling high-performance data updates within the applications architecture<br>- It leverages PostgreSQLs copy protocol to handle large datasets seamlessly, supporting the overall system's need for scalable and rapid leaderboard data management<br>- This component is integral to maintaining accurate, real-time rankings in the platform.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/leaderboard.sql.go'>leaderboard.sql.go</a></b></td>
									<td style='padding: 8px;'>- Leaderboard SQL Query HandlerThis code file provides the core logic for generating comprehensive user rankings based on their ride activity<br>- Specifically, it calculates the all-time leaderboard by aggregating total distance traveled and ride count for each user with completed rides<br>- This functionality supports features such as displaying the top users in the application, fostering competition, and encouraging engagement by showcasing cumulative performance metrics across the entire user base<br>- It plays a crucial role within the broader architecture by enabling efficient retrieval of aggregated user statistics for leaderboard displays and related analytics.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/sqlc/models.go'>models.go</a></b></td>
									<td style='padding: 8px;'>- Defines data models representing core entities such as rides, users, vehicles, and related relationships within the applications database schema<br>- Facilitates structured data access, manipulation, and serialization, supporting features like ride tracking, user profiles, vehicle management, and leaderboards, thereby underpinning the applications data layer and ensuring consistency across the system architecture.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- queries Submodule -->
					<details>
						<summary><b>queries</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ internal.db.queries</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/queries/vehicles.sql'>vehicles.sql</a></b></td>
									<td style='padding: 8px;'>- Defines database operations for managing vehicle records, including creation, retrieval, updating, deactivation, and deletion, while ensuring data integrity and consistency<br>- Supports querying vehicles by user, checking for active rides, and enforcing constraints during deletions<br>- Facilitates seamless integration of vehicle data within the broader application architecture, enabling efficient vehicle lifecycle management.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/queries/leaderboard.sql'>leaderboard.sql</a></b></td>
									<td style='padding: 8px;'>- Defines SQL queries for managing and retrieving leaderboard data across various periods and vehicle types<br>- Facilitates insertion, deletion, and fetching of user rankings, statistics, and friend-based leaderboards, supporting the core functionality of tracking and displaying user performance metrics within the applications competitive ecosystem.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/queries/auth.sql'>auth.sql</a></b></td>
									<td style='padding: 8px;'>- Defines SQL queries for managing user data within the authentication subsystem, enabling creation, retrieval, and updates of user profiles<br>- These queries facilitate core user operations such as registration, profile viewing, and push notification token management, supporting seamless user account handling and data consistency across the applications architecture.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/queries/social.sql'>social.sql</a></b></td>
									<td style='padding: 8px;'>- Defines SQL queries for managing social interactions, including following/unfollowing users, liking rides, commenting, and retrieving personalized activity feeds<br>- Facilitates core social features within the platform, enabling users to connect, share, and engage with ride content from followed users, thereby supporting the applications social and community-building architecture.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/db/queries/rides.sql'>rides.sql</a></b></td>
									<td style='padding: 8px;'>- Defines database queries for managing ride lifecycle, including creation, retrieval, updating, and abandonment, as well as tracking GPS points associated with each ride<br>- Facilitates efficient storage, access, and analysis of ride data, supporting features like ride history, real-time tracking, and statistical summaries within the overall system architecture.</td>
								</tr>
							</table>
						</blockquote>
					</details>
				</blockquote>
			</details>
			<!-- websocket Submodule -->
			<details>
				<summary><b>websocket</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.websocket</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/websocket/buffer.go'>buffer.go</a></b></td>
							<td style='padding: 8px;'>- Implements a GPS data batching and buffering system to efficiently collect, manage, and persist real-time GPS points for active rides<br>- It ensures reliable, high-performance database updates through concurrent connection management, periodic flushing, and batch inserts, supporting scalable ride tracking within the overall web socket-based architecture.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/websocket/hub.go'>hub.go</a></b></td>
							<td style='padding: 8px;'>- Manages WebSocket connections for real-time GPS data streaming within the application architecture<br>- Facilitates secure connection upgrades, validates tokens, and handles incoming GPS points, ensuring efficient data flow from clients to the GPS buffer<br>- Integrates with Redis for authentication and session management, supporting live tracking features across the system.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- handler Submodule -->
			<details>
				<summary><b>handler</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.handler</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/users.go'>users.go</a></b></td>
							<td style='padding: 8px;'>- Provides user profile management within the application, enabling retrieval and updates of authenticated user data, as well as access to public profiles by ID<br>- Integrates with the database to ensure data consistency and security, supporting seamless user profile interactions aligned with the overall system architecture.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/auth.go'>auth.go</a></b></td>
							<td style='padding: 8px;'>- Defines authentication endpoints for user registration, login, token refresh, and logout within the application<br>- Facilitates secure user identity management by handling credential validation, token issuance, and session control, integrating with core services and database queries to support the overall architectures authentication layer.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/social.go'>social.go</a></b></td>
							<td style='padding: 8px;'>- Facilitates social interactions within the platform by managing user follow/unfollow actions, ride likes, and comments, while also providing a personalized feed of rides<br>- Ensures seamless user engagement through notifications and maintains data consistency across user relationships and content interactions, integrating core functionalities essential for fostering community and activity within the application.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/helpers.go'>helpers.go</a></b></td>
							<td style='padding: 8px;'>- Provides utility functions for request handling, including parsing UUIDs, managing optional string values, and formatting error responses<br>- Facilitates consistent error reporting and data validation within the web applications API layer, supporting robust and user-friendly interactions across the codebase.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/handlers.go'>handlers.go</a></b></td>
							<td style='padding: 8px;'>- Defines HTTP handlers for core application functionalities, including user authentication, vehicle management, ride operations, social interactions, leaderboards, and health checks<br>- Serves as the central routing layer, orchestrating request handling and integrating services with database and cache layers to facilitate seamless API interactions within the overall architecture.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/rides.go'>rides.go</a></b></td>
							<td style='padding: 8px;'>- Defines HTTP handlers for managing ride lifecycle operations, including starting, stopping, listing, and retrieving rides<br>- Facilitates user authentication, input validation, and interaction with backend services to coordinate ride data, while also handling real-time WebSocket token management<br>- Integrates seamlessly into the overall architecture to enable secure, efficient ride management within the application.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/leaderboard.go'>leaderboard.go</a></b></td>
							<td style='padding: 8px;'>- Provides API endpoints to retrieve global and friends-specific leaderboards, supporting filtering by period and vehicle type<br>- Facilitates ranking users based on ride metrics within defined timeframes, enabling competitive insights and social engagement<br>- Ensures accurate period calculations aligned with timezone settings, integrating seamlessly into the overall ride tracking and social features of the platform.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/handler/vehicles.go'>vehicles.go</a></b></td>
							<td style='padding: 8px;'>- Defines HTTP handlers for managing vehicle data within the application, enabling users to create, retrieve, update, and delete their vehicles<br>- Ensures proper authorization, validation, and data consistency, integrating with the database to support vehicle lifecycle operations while maintaining data integrity and handling edge cases like active ride associations.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- service Submodule -->
			<details>
				<summary><b>service</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.service</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/service/auth.go'>auth.go</a></b></td>
							<td style='padding: 8px;'>- Provides core authentication functionalities, including user registration, login, password management, and token generation within the applications architecture<br>- Facilitates secure user identity verification and session management through JWT tokens, ensuring robust access control and seamless user authentication workflows across the system.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/service/rides.go'>rides.go</a></b></td>
							<td style='padding: 8px;'>- Provides core ride management functionalities, including initiating, stopping, retrieving, and listing rides within the application<br>- Ensures proper validation, ownership verification, and metrics calculation based on GPS data, supporting the overall architecture of user-specific ride tracking, data integrity, and performance monitoring in the system.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/service/notifications.go'>notifications.go</a></b></td>
							<td style='padding: 8px;'>- Implements a notification service that facilitates sending real-time push notifications via the Expo Push API, specifically for user interactions such as likes and comments on rides<br>- It integrates seamlessly into the applications architecture to enhance user engagement by delivering timely alerts based on user activity.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/service/metrics.go'>metrics.go</a></b></td>
							<td style='padding: 8px;'>- Provides core functionality for calculating and summarizing ride metrics, including distance, duration, speed, elevation gain, and calories, based on GPS data<br>- Facilitates route visualization through polyline encoding and bounding box computation, supporting overall architecture by enabling detailed ride analysis and efficient route representation within the system.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/service/placeholder.go'>placeholder.go</a></b></td>
							<td style='padding: 8px;'>- Defines the core business logic services across key functionalities such as authentication, ride metrics, social interactions, notifications, and geocoding within the application architecture<br>- Serves as a foundational placeholder to organize and integrate these services, ensuring modularity and scalability in the overall system design<br>- Establishes a versioning reference for service consistency and future development.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/service/metrics_test.go'>metrics_test.go</a></b></td>
							<td style='padding: 8px;'>- Provides unit tests for core metrics computation functions within the ride tracking system<br>- Validates calculations of distance, duration, speed, elevation gain, and calories based on GPS points, ensuring accuracy across various scenarios<br>- Supports the overall architecture by verifying the correctness of data processing and analytical components essential for ride summaries and performance insights.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- config Submodule -->
			<details>
				<summary><b>config</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.config</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/config/config.go'>config.go</a></b></td>
							<td style='padding: 8px;'>- Defines and loads application configuration settings by reading environment variables, ensuring proper validation and parsing of critical parameters such as database connections, JWT secrets, and external API keys<br>- Serves as the central configuration module, enabling consistent and secure setup across the entire codebase, facilitating environment-specific adjustments and streamlined initialization.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- router Submodule -->
			<details>
				<summary><b>router</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.router</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/router/router.go'>router.go</a></b></td>
							<td style='padding: 8px;'>- Defines and configures the API routing structure for the application, organizing endpoints for health checks, authentication, user management, vehicle operations, ride tracking, social interactions, leaderboards, and WebSocket streaming<br>- Ensures proper middleware application for security and cross-origin requests, facilitating seamless communication between clients and backend services within the overall system architecture.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- middleware Submodule -->
			<details>
				<summary><b>middleware</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.middleware</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/middleware/auth.go'>auth.go</a></b></td>
							<td style='padding: 8px;'>- Implements JWT-based authentication middleware for Gin, ensuring secure access control by validating tokens, extracting user identifiers, and making user data accessible to downstream handlers<br>- Facilitates consistent authorization checks across the application, supporting a secure and streamlined user authentication flow within the overall system architecture.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/middleware/cors.go'>cors.go</a></b></td>
							<td style='padding: 8px;'>- Implements Cross-Origin Resource Sharing (CORS) middleware to enable secure cross-origin requests across the application<br>- It manages HTTP headers to control access from different origins, facilitating smooth client-server interactions during development and production<br>- This middleware ensures that the API can handle cross-origin requests appropriately, maintaining security and functionality within the overall architecture.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- errors Submodule -->
			<details>
				<summary><b>errors</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.errors</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/errors/errors.go'>errors.go</a></b></td>
							<td style='padding: 8px;'>- Defines a centralized error handling framework with domain-specific error types and constants, facilitating consistent error representation and HTTP status mapping across the application<br>- Supports clear identification, creation, and categorization of errors related to authentication, vehicle management, rides, validation, and general server issues, thereby enhancing robustness and maintainability within the overall architecture.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- jobs Submodule -->
			<details>
				<summary><b>jobs</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ internal.jobs</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/internal/jobs/leaderboard.go'>leaderboard.go</a></b></td>
							<td style='padding: 8px;'>- Implements scheduled background jobs to compute and update weekly and monthly leaderboard rankings across all users and specific vehicle types<br>- Ensures periodic recalculations, data consistency, and efficient bulk insertions, supporting the overall architecture of user engagement and performance tracking within the application.</td>
						</tr>
					</table>
				</blockquote>
			</details>
		</blockquote>
	</details>
	<!-- cmd Submodule -->
	<details>
		<summary><b>cmd</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ cmd</b></code>
			<!-- server Submodule -->
			<details>
				<summary><b>server</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ cmd.server</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/cmd/server/main.go'>main.go</a></b></td>
							<td style='padding: 8px;'>- Orchestrates the startup, configuration, and graceful shutdown of the backend server, integrating database connections, caching, WebSocket communication, and scheduled jobs<br>- Ensures reliable initialization of core services, manages server lifecycle events, and coordinates cleanup procedures, thereby maintaining the overall stability and responsiveness of the entire application architecture.</td>
						</tr>
					</table>
				</blockquote>
			</details>
		</blockquote>
	</details>
	<!-- pkg Submodule -->
	<details>
		<summary><b>pkg</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ pkg</b></code>
			<!-- jwt Submodule -->
			<details>
				<summary><b>jwt</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ pkg.jwt</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/pkg/jwt/jwt_test.go'>jwt_test.go</a></b></td>
							<td style='padding: 8px;'>- Provides comprehensive tests for JWT token generation and validation, ensuring secure authentication workflows within the overall architecture<br>- Validates token creation, parsing, expiration handling, signature verification, and claim extraction, thereby safeguarding the integrity and reliability of the authentication mechanism across the system.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/pkg/jwt/jwt.go'>jwt.go</a></b></td>
							<td style='padding: 8px;'>- Provides core JWT functionality for token creation, validation, and parsing within the authentication system<br>- It supports generating access and refresh tokens with custom claims, validating token integrity, and extracting user identifiers, thereby enabling secure user session management and seamless authentication workflows across the application.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- polyline Submodule -->
			<details>
				<summary><b>polyline</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ pkg.polyline</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/pkg/polyline/polyline_test.go'>polyline_test.go</a></b></td>
							<td style='padding: 8px;'>- Provides comprehensive testing for polyline encoding and decoding functionalities, ensuring accurate transformation of geographic coordinate sequences into compact string representations and vice versa<br>- Validates handling of edge cases, negative and large coordinates, and precision preservation, thereby guaranteeing robustness and reliability within the spatial data processing architecture.</td>
						</tr>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='https://github.com/nashirabbash/myRiders-BE/blob/master/pkg/polyline/polyline.go'>polyline.go</a></b></td>
							<td style='padding: 8px;'>- Provides encoding and decoding functionalities for Google Encoded Polyline format, enabling efficient representation and retrieval of geographic coordinate sequences<br>- Facilitates seamless integration of route data within mapping applications by translating between raw coordinate arrays and compact string representations, supporting optimized storage and transmission in the overall geospatial architecture.</td>
						</tr>
					</table>
				</blockquote>
			</details>
		</blockquote>
	</details>
</details>

---

## Getting Started

### Prerequisites

This project requires the following dependencies:

- **Programming Language:** Go
- **Package Manager:** Go modules

### Installation

Build myRiders-BE from the source and install dependencies:

1. **Clone the repository:**

   ```sh
   ❯ git clone https://github.com/nashirabbash/myRiders-BE
   ```

2. **Navigate to the project directory:**

   ```sh
   ❯ cd myRiders-BE
   ```

3. **Install the dependencies:**

**Using [go modules](https://golang.org/):**

```sh
❯ go build
```

### Usage

Run the project with:

**Using [go modules](https://golang.org/):**

```sh
go run {entrypoint}
```

### Testing

Myriders-be uses the {**test_framework**} test framework. Run the test suite with:

**Using [go modules](https://golang.org/):**

```sh
go test ./...
```

---

## Roadmap

- [x] **`Task 1`**: <strike>Implement feature one.</strike>
- [ ] **`Task 2`**: Implement feature two.
- [ ] **`Task 3`**: Implement feature three.

---

## Contributing

- **💬 [Join the Discussions](https://github.com/nashirabbash/myRiders-BE/discussions)**: Share your insights, provide feedback, or ask questions.
- **🐛 [Report Issues](https://github.com/nashirabbash/myRiders-BE/issues)**: Submit bugs found or log feature requests for the `myRiders-BE` project.
- **💡 [Submit Pull Requests](https://github.com/nashirabbash/myRiders-BE/blob/main/CONTRIBUTING.md)**: Review open PRs, and submit your own PRs.

<details closed>
<summary>Contributing Guidelines</summary>

1. **Fork the Repository**: Start by forking the project repository to your github account.
2. **Clone Locally**: Clone the forked repository to your local machine using a git client.
   ```sh
   git clone https://github.com/nashirabbash/myRiders-BE
   ```
3. **Create a New Branch**: Always work on a new branch, giving it a descriptive name.
   ```sh
   git checkout -b new-feature-x
   ```
4. **Make Your Changes**: Develop and test your changes locally.
5. **Commit Your Changes**: Commit with a clear message describing your updates.
   ```sh
   git commit -m 'Implemented new feature x.'
   ```
6. **Push to github**: Push the changes to your forked repository.
   ```sh
   git push origin new-feature-x
   ```
7. **Submit a Pull Request**: Create a PR against the original project repository. Clearly describe the changes and their motivations.
8. **Review**: Once your PR is reviewed and approved, it will be merged into the main branch. Congratulations on your contribution!
</details>

<details closed>
<summary>Contributor Graph</summary>
<br>
<p align="left">
   <a href="https://github.com{/nashirabbash/myRiders-BE/}graphs/contributors">
      <img src="https://contrib.rocks/image?repo=nashirabbash/myRiders-BE">
   </a>
</p>
</details>

---

## License

Myriders-be is protected under the [LICENSE](https://choosealicense.com/licenses) License. For more details, refer to the [LICENSE](https://choosealicense.com/licenses/) file.

---

## Acknowledgments

- Credit `contributors`, `inspiration`, `references`, etc.

<div align="left"><a href="#top">⬆ Return</a></div>

---
