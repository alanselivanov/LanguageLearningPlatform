# Language Learning Platform

<<<<<<< HEAD
## Project Overview


**Language Learning Platform** is a web-based application designed to facilitate the learning of languages. The platform allows users to engage in basic language learning exercises, with functionalities such as user registration, authentication, and progress tracking.

## Key Features
- **Database Integration**: The application is connected to a PostgreSQL database via pgAdmin4 for efficient data management.
- **CRUD Operations**: Admins and users can perform basic create, read, update, and delete operations on their data.
- **Interactive Interface**: A simple, responsive interface for interacting with the platform’s functionalities.

## Target Audience
This platform is intended for individuals interested in learning a new language or improving their skills. It caters to both beginners and more experienced learners who want a straightforward tool to help track their progress.

## Team Members
- **Selivanov Alan** 
- **Khusainov Almas**
- **Baltabayev Adil**

## Getting Started
Follow these instructions to get the Language Learning Platform running on your local machine.

### Prerequisites
- Go programming language installed on your machine.
- PostgreSQL installed for managing databases.
- Go packages: `github.com/lib/pq` and `gorm.io/gorm`.

### Step-by-Step Guide
1. **Clone the repository**:
    ```bash
    git clone https://github.com/alanselivanov/LanguageLearningPlatform.git
    cd LanguageLearningPlatform
    ```

2. **Install Go dependencies**: 
   Install the required Go packages:
    ```bash
    go get -u github.com/lib/pq
    go get -u gorm.io/gorm
    ```

3. **Set Up PostgreSQL Database**:
   - Create a PostgreSQL database and user for your project.
   - Set up the necessary tables, such as the `users` table with fields like `id`, `name`, `email`, `created_at`, and `updated_at`.

4. **Run the server**: 
   Start the Go server:
    ```bash
    go run main.go
    ```
   The server will run on port 8080.

5. **Access the platform**: 
   Open your browser and go to `http://localhost:8080` to access the Language Learning Platform.

## Tools and Technologies Used
- **Go (Golang)**: Backend programming language used to build the server.
- **PostgreSQL**: Database used to store user data and course information.
- **GORM**: ORM (Object-Relational Mapping) library used to interact with the PostgreSQL database.
- **HTML/CSS/JavaScript**: Frontend technologies used for the user interface.
- **Postman**: Used for testing the API endpoints.

## Future Enhancements
- Extend the platform with additional features like quizzes, progress tracking, and personalized learning goals.
- Enhance the UI for a more engaging and user-friendly experience.
=======
>>>>>>> b7f04e82dac0a3753a3fc8c57c08c889fd59c2a8
