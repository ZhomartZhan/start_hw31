package start_hw31

import (
	"encoding/json"
	redis_lib "github.com/ZhomartZhan/common_lib_hw31"
	users "github.com/ZhomartZhan/users_hw31"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpEndpoints interface {
	TestEndpoint() func(w http.ResponseWriter, r *http.Request)
	TestEndpointWithParam(idParam string) func(w http.ResponseWriter, r *http.Request)
	TestPostEndpoint() func(w http.ResponseWriter, r *http.Request)
	RegisterEndpoint() func(w http.ResponseWriter, r *http.Request)
	LoginEndpoint() func(w http.ResponseWriter, r *http.Request)
	ProfileEndpoint() func(w http.ResponseWriter, r *http.Request)
}

type httpEndpoints struct {
	//variable connection to db
	usersStore users.UsersStore
	redisStore *redis_lib.RedisStore
}

func NewHttpEndpoints(uS users.UsersStore, rS *redis_lib.RedisStore) HttpEndpoints {
	return &httpEndpoints{usersStore: uS, redisStore: rS}
}

func (h *httpEndpoints) TestEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user := users.User{
			Id:        "100513974",
			Username:  "TestUsername1",
			Password:  "Qwerty11!",
			FirstName: "Dana",
			LastName:  "White",
			Avatar:    "picture1",
		}
		respondJSON(w, http.StatusOK, user)
		return
	}
}

func (h *httpEndpoints) TestEndpointWithParam(idParam string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, ok := vars[idParam]
		if !ok {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    "Dont have user with that id",
				StatusCode: http.StatusBadRequest,
			})
		}
		var response users.User
		usersData := []users.User{
			{
				Id:        "1",
				Username:  "Cool_Dude",
				Password:  "qweasdzxc",
				FirstName: "Alex",
				LastName:  "Hopkins",
				Avatar:    "picture2",
			},
			{
				Id:        "2",
				Username:  "FeelsBadMan",
				Password:  "rtyfghvbn",
				FirstName: "Jack",
				LastName:  "Smith",
				Avatar:    "picture3",
			},
		}
		if idStr == "1" {
			response = usersData[0]
		} else if idStr == "2" {
			response = usersData[1]
		} else {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    "Dont have user with that id",
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		respondJSON(w, http.StatusOK, response)
		return
	}
}

func (h *httpEndpoints) TestPostEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		user := &users.User{}
		err = json.Unmarshal(jsonData, &user)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		user.Id = "3333"
		respondJSON(w, http.StatusCreated, user)
		return
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func (h *httpEndpoints) RegisterEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		user := &users.User{}
		err = json.Unmarshal(jsonData, &user)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		if user.Username == "" || user.Password == "" {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    ErrUsernamePasswordEmpty.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		oldUser, err := h.usersStore.GetByUsernameAndPassword(user.Username, user.Password)
		if err != nil && err != users.ErrNoUser {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		if oldUser != nil {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    ErrUserAlreadyExist.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		response, err := h.usersStore.Create(user)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		respondJSON(w, http.StatusCreated, response)
		return
	}
}

func (h *httpEndpoints) LoginEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		req := &LoginRequest{}
		err = json.Unmarshal(jsonData, &req)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		user, err := h.usersStore.GetByUsernameAndPassword(req.Username, req.Password)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		key := uuid.New().String()
		err = h.redisStore.SetValue(key, user, 5*time.Minute)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		response := &LoginResponse{AccessKey: key}
		respondJSON(w, http.StatusOK, response)
		return
	}
}

func (h *httpEndpoints) ProfileEndpoint() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		contextData := r.Context().Value("user_id")
		userId := contextData.(string)
		response, err := h.usersStore.Get(userId)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, HttpError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		respondJSON(w, http.StatusOK, response)
		return
	}
}
