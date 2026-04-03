
### Technology Stack Selection:
- Backend: Go (Golang) 
Chosen for its simplicity, strongly typed struckture, concurrency capabilities, and high performance in backend services.
- Framework: Gin 
Lightweight and fast HTTP framework that simplifies routing and middleware handling.
- Database: PostgreSQL
Reliable, open-source relational database with strong consistency.
- ORM: GORM
It is the most popular Go ORM so it properly maintained by the community.
- Testing tools: testify
Provides expressive assertions and testing utilities, improving test readability and maintainability.
- Testing tools: mockery
Simplifies mock generation, enabling clean and isolated unit testing.


### Clean Architecture: 

Promotes separation of concerns, making the system easier to maintain, test, and extend. The system is organized into distinct layers with clear responsibilities:

- API Layer (Handlers/Controllers): Handles HTTP requests, input validation, and response formatting.
- Application Layer (Use Cases): Orchestrates business workflows and coordinates between domain logic and external services.
- Domain Layer (Entities & Interfaces): Contains core business logic, entities, and contracts, remaining independent from frameworks and infrastructure.
- Infrastructure Layer (Persistence & External Services): Implements database access, external APIs, and other technical details.