┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                                                                          Casino Transaction Management System                                                                          ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
                                                                                                                                                                                          
                                                                                                                                                                                          
                                                                                         Overview                                                                                         
                                                                                                                                                                                          
This project implements a simple transaction management system for a casino. It tracks user transactions related to bets and wins, processes them asynchronously via RabbitMQ, stores data
in PostgreSQL, and exposes a REST API for querying transaction data.                                                                                                                      
                                                                                                                                                                                          
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
                                                                                                                                                                                          
                                                                                        Components                                                                                        
                                                                                                                                                                                          
 • Message System: RabbitMQ is used to receive and process bet/win transaction messages asynchronously.                                                                                   
 • Database: PostgreSQL stores transaction data with fields: user_id, transaction_type (bet or win), amount, and timestamp.                                                               
 • API: A Go-based REST API allows querying transactions with filtering by user and transaction type.                                                                                     
 • Consumer: A Go service consumes messages from RabbitMQ and saves transactions to the database.                                                                                         

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

                                                                                      Setup and Run                                                                                       

 1 Navigate to the docker directory:          

                                                                                                                                                                                          
 cd docker                                                                                                                                                                                
                                                                                                                                                                                          

 2 For normal application run:                

 • Ensure the postgres_test service is commented out or removed in docker-compose.yml.       
 • Start the main services:                   

                                                                                                                                                                                          
 docker-compose up -d api consumer postgres rabbitmq                                                                                                                                      
                                                                                                                                                                                          

 3 For integration testing:                   

 • Uncomment the postgres_test service in docker-compose.yml.                                
 • Start the test database and RabbitMQ:      

                                                                                                                                                                                          
 docker-compose up -d postgres_test rabbitmq                                                                                                                                              
                                                                                                                                                                                          

 • Set environment variables for tests to connect to the test database and RabbitMQ:         

                                                                                                                                                                                          
 export TEST_DB_USER=postgres                                                                                                                                                             
 export TEST_DB_PASSWORD=postgres                                                                                                                                                         
 export TEST_DB_HOST=localhost                                                                                                                                                            
 export TEST_DB_PORT=5433                                                                                                                                                                 
 export TEST_DB_NAME=casino_test                                                                                                                                                          
 export TEST_RABBITMQ_URL=amqp://guest:guest@localhost:5672/                                                                                                                              
 export TEST_RABBITMQ_QUEUE=transactions                                                                                                                                                  
                                                                                                                                                                                          

 • Run tests:
                                                                                                                                                                                           
 go test ./tests -v                                                                                                                                                                       
                                                                                                                                                                                          

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

                                                                                          Usage                                                                                           

                                                                                   Sending Transactions                                                                                   

 • Use RabbitMQ UI or any AMQP client to publish messages to the transactions queue.                                                                                                      
 • Message format (JSON):                                                                                                                                                                 

                                                                                                                                                                                          
 {                                                                                                                                                                                        
   "user_id": 1,                                                                                                                                                                          
   "transaction_type": "bet",                                                                                                                                                             
   "amount": 100.0,                                                                                                                                                                       
   "timestamp": "2024-01-01T12:00:00Z"                                                                                                                                                    
 } 

 Querying Transactions                                                                                   

 • Get all transactions:                      

                                                                                                                                                                                          
 curl http://localhost:8080/transactions                                                                                                                                                  
                                                                                                                                                                                          

 • Filter by user:                            

                                                                                                                                                                                          
 curl "http://localhost:8080/transactions?user_id=1"                                                                                                                                      
                                                                                                                                                                                          

 • Filter by transaction type:                

                                                                                                                                                                                          
 curl "http://localhost:8080/transactions?transaction_type=win"                                                                                                                           
                                                                                                                                                                                          

 • Combined filter:                           

                                                                                                                                                                                          
 curl "http://localhost:8080/transactions?user_id=1&transaction_type=bet"                                                                                                                 
                                                                                                                                                                                          

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

                                                                                         Testing                                                                                          

 • Unit and integration tests are located in the tests/ directory.                           
 • Integration tests require the test PostgreSQL database (postgres_test) and RabbitMQ to be running.                                                                                     
 • Run all tests with:                        

                                                                                                                                                                                          
 go test ./tests -v                                                                                                                                                                       
                                                                                                                                                                                          

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

                                                                                    Stopping Services

To stop all running containers:                                                                                                                                                           

                                                                                                                                                                                          
 docker-compose down                                                                                                                                                                      
                                                                                                                                                                                          

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────

                                                                                     Additional Notes                                                                                     

 • Always run Docker Compose commands from the docker directory to ensure correct paths.                                                                                                  
 • For integration tests, ensure environment variables point to the test database and RabbitMQ.                                                                                           
 • The postgres_test service should be disabled during normal runs to avoid conflicts.