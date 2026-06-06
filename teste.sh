# ==========================================
# CADASTRO DE UTILIZADOR
# ==========================================
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bob@email.com",
    "password": "senha123"
  }' | jq

# ==========================================
# LOGIN
# ==========================================
curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bob@email.com",
    "password": "senha123",
    "expires_in_seconds": 3600
  }' | jq

# ==========================================
# LOGIN (sem expires_in_seconds — usa 3600 por defeito)
# ==========================================
curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bob@email.com",
    "password": "senha123"
  }' | jq

# ==========================================
# CRIAR CHIRP (substituir TOKEN e USER_ID pelos valores reais)
# ==========================================
curl -s -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN_AQUI" \
  -d '{
    "body": "Olá mundo!",
    "user_id": "USER_ID_AQUI"
  }' | jq

# ==========================================
# LISTAR TODOS OS CHIRPS
# ==========================================
curl -s http://localhost:8080/api/chirps | jq

# ==========================================
# BUSCAR CHIRP POR ID
# ==========================================
curl -s http://localhost:8080/api/chirps/CHIRP_ID_AQUI | jq