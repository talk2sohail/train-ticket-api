# train-ticket-api

A gRPC based microservice for train ticket booking.

## Folder Structure

- **/cmd**  
  An example application that uses the services APIs

  - _main.go_: Starts the application, initializes the gRPC client and demonstrates sample service interactions.

- **/client**

  NOTE: Currently, the client code is tightly coupled with gRPC client code. It has been kept that way to keep things simple for now. We can decouple the gRPC code to make it more readable, stable and user-friendly.

  Implements the client-side code to interact with the gRPC server.

  - _ticket_client.go_: Handles connection setup, RPC calls, and abstracts communication with the ticketing service.

- **/internal**  
  Holds the core business logic and implementation details.

  - **/internal/common**  
    Contains shared components used across the project.
    - _genproto_: Auto-generated code from Protocol Buffers defining messages and services.
    - _util_: Utility functions such as validation helpers and common error handling.
  - **/internal/ticket**  
    Implements the ticketing functionality.
    - **/internal/ticket/service**  
      Contains the business logic for ticket purchase, seat allocation etc
    - **/internal/ticket/handler**  
      Maps incoming gRPC requests to the service layer and handles protocol-specific operations.
    - _service_test.go_ and _handler_test.go_: Unit tests ensuring the reliability of services and handlers.

- **/proto**  
  Includes the original Protocol Buffer (.proto) files that define the service contracts and messages.

## Features

- **Purchase Ticket**:  
  Facilitates ticket booking by allocating the first available seat from defined sections and generating a unique ticket receipt.

- **Receipt Generation**:  
  Automatically produces a detailed receipt containing ticket ID, journey details, user information, and purchase timestamp.

- **Modify Seat**:  
  Allows users to change their seat allocation if the desired seat is available, updating the booking accordingly.

- **Remove User**:  
  Supports cancellation by removing a user's booking, thereby freeing up the occupied seat for future bookings.

- **Get Receipt Details**:  
  Retrieves detailed booking information for a given ticket ID, aiding in user queries and support.

- **Get Users by Section**:  
  Lists users and their allocated seats for a specific section, useful for monitoring seat occupancy and service analytics.

## Areas for Improvement

- **Enhanced Error Handling**:  
  Improve error propagation and logging. Consider standardized error responses, better error categorization, and integration with monitoring tools.

- **Client API Support**:  
  Expand and decouple the client API to simplify integration. Future improvements might involve creating a more generic, RESTful interface or better abstractions for gRPC interactions.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributors

- Md Sohail [Github](https://github.com/talk2sohail)
