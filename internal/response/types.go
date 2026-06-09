
package response

type ChirpResponse struct {
	Id         string `json:"id"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
	Body       string `json:"body"`
	User_id    string `json:"user_id"`
}


type UserResponse struct {
	Id            string `json:"id"`
	Created_at    string `json:"created_at"`
	Updated_at    string `json:"updated_at"`
	Email         string `json:"email"`
	Is_chirpy_red bool   `json:"is_chirpy_red"`
}