package keycloak

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"path"
	"strconv"
)

func (c *client) GetUser(id string) (User, error) {
	user := User{}
	if len(id) == 0 {
		return user, errors.New("user 'id' must be set")
	}

	status, body, err := fiber.Get(c.address+"/admin/realms/"+path.Join(c.realm, "users", id)).
		Set("Authorization", "Bearer "+c.getToken()).
		JSONDecoder(json.Unmarshal).
		Struct(&user)
	if status != 200 || len(err) > 0 {
		c.logger.Error("filed to get user", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		return user, errors.Join(err...)
	}

	return user, nil
}

func (c *client) CreateUser(user User) error {
	user.Id = ""
	status, body, err := fiber.Post(c.address+"/admin/realms/"+path.Join(c.realm, "users")).
		Set("Authorization", "Bearer "+c.getToken()).
		JSONEncoder(json.Marshal).
		JSON(&user).Bytes()
	if status != 201 || len(err) > 0 {
		c.logger.Error("filed to create user", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		return errors.Join(err...)
	}
	return nil
}

func (c *client) UpdateUser(user User) error {
	if len(user.Id) == 0 {
		return errors.New("user 'id' must be set")
	}
	status, body, err := fiber.Put(c.address+"/admin/realms/"+path.Join(c.realm, "users", user.Id)).
		JSONEncoder(json.Marshal).
		Set("Authorization", "Bearer "+c.getToken()).JSON(&user).Bytes()

	if status != 204 || len(err) > 0 {
		c.logger.Error("filed to update user", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		if status != 204 {
			err = append(err, errors.New(""+strconv.FormatInt(int64(status), 10)))
		}
		return errors.Join(err...)
	}
	return nil
}

func (c *client) DeleteUser(id string) error {
	if len(id) == 0 {
		return errors.New("user 'id' must be set")
	}

	status, body, err := fiber.Delete(c.address+"/admin/realms/"+path.Join(c.realm, "users", id)).
		Set("Authorization", "Bearer "+c.getToken()).Bytes()

	if len(err) > 0 {
		c.logger.Error("filed to delete user", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		return errors.Join(err...)
	}
	return nil
}

func (c *client) GetUserByUsername(username string) (User, error) {
	if len(username) == 0 {
		return User{}, errors.New("'username' must be set")
	}

	users := []User{}
	status, body, err := fiber.Get(c.address+"/admin/realms/"+path.Join(c.realm, "users")).
		JSONDecoder(json.Unmarshal).
		QueryString("exact=true&username="+username).
		Set("Authorization", "Bearer "+c.getToken()).Struct(&users)
	if status != 200 || len(err) > 0 {
		c.logger.Error("filed to get user by username", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		return User{}, errors.Join(err...)
	}

	if len(users) == 0 {
		return User{}, errors.New("user not found")
	}

	return users[0], nil
}

func (c *client) getClientRoles(clientId string) ([]Role, error) {
	roles := []Role{}
	if len(clientId) == 0 {
		return nil, errors.New("'clientId' must be set")
	}

	status, body, err := fiber.Get(c.address+"/admin/realms/"+path.Join(c.realm, "clients", clientId, "roles")).
		JSONDecoder(json.Unmarshal).
		Set("Authorization", "Bearer "+c.getToken()).
		Struct(&roles)

	if status != 200 || len(err) > 0 {
		c.logger.Error("filed to get client roles", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		if status != 200 {
			err = append(err, errors.New(strconv.FormatInt(int64(status), 10)))
		}
		return nil, errors.Join(err...)
	}

	return roles, nil
}

func (c *client) getClient(clientName string) (ClientRep, error) {
	cl := ClientRep{}
	if len(clientName) == 0 {
		return cl, errors.New("'clientId' must be set")
	}

	clients := []ClientRep{}
	status, body, err := fiber.Get(c.address+"/admin/realms/"+path.Join(c.realm, "clients")).
		JSONDecoder(json.Unmarshal).
		QueryString("clientId="+clientName).
		Set("Authorization", "Bearer "+c.getToken()).
		Struct(&clients)
	if status != 200 || len(err) > 0 {
		c.logger.Error("filed to get client", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		if status != 200 {
			err = append(err, errors.New(strconv.FormatInt(int64(status), 10)))
		}
		return cl, errors.Join(err...)
	}

	return clients[0], nil
}

func (c *client) addUserClientRoles(id string, clientId string, roles []Role) error {
	if len(id) == 0 || len(clientId) == 0 {
		return errors.New("user 'id' and 'clientId' must be set")
	} else if len(roles) == 0 {
		return errors.New("roles len == 0, if you need remove roles use: DeleteUserClientRoles")
	}

	status, body, err := fiber.Post(c.address+"/admin/realms/"+path.Join(c.realm, "users", id, "role-mappings/clients", clientId)).
		JSONEncoder(json.Marshal).
		Set("Authorization", "Bearer "+c.getToken()).
		JSON(&roles).Bytes()
	if status != 204 || len(err) > 0 {
		c.logger.Error("filed to add user client roles", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		if status != 204 {
			err = append(err, errors.New(""+strconv.FormatInt(int64(status), 10)))
		}
		return errors.Join(err...)
	}

	return nil
}

func (c *client) getUserClientRoles(id string, clientId string) ([]Role, error) {
	if len(id) == 0 || len(clientId) == 0 {
		return nil, errors.New("user 'id', 'clientId' must be set")
	}

	roles := []Role{}
	status, body, err := fiber.Get(c.address+"/admin/realms/"+path.Join(c.realm, "users", id, "role-mappings/clients", clientId)).
		JSONDecoder(json.Unmarshal).
		Set("Authorization", "Bearer "+c.getToken()).
		Struct(&roles)
	if status != 200 || len(err) > 0 {
		c.logger.Error("filed to get user client roles", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		if status != 200 {
			err = append(err, errors.New(""+strconv.FormatInt(int64(status), 10)))
		}
		return nil, errors.Join(err...)
	}

	return roles, nil
}

func (c *client) deleteUserClientRoles(id string, clientId string, roles []Role) error {
	if len(id) == 0 || len(clientId) == 0 || len(roles) == 0 {
		return errors.New("user 'id' and 'clientId' and 'roles' must be set")
	}

	status, body, err := fiber.Delete(c.address+"/admin/realms/"+path.Join(c.realm, "users", id, "role-mappings/clients", clientId)).
		JSONEncoder(json.Marshal).
		Set("Authorization", "Bearer "+c.getToken()).
		JSON(&roles).Bytes()
	if status != 204 || len(err) > 0 {
		c.logger.Error("filed to delete user client roles", zap.String("body", string(body)), zap.Int("status", status), zap.Errors("errors", err))
		if status != 204 {
			err = append(err, errors.New(""+strconv.FormatInt(int64(status), 10)))
		}
		return errors.Join(err...)
	}

	return nil
}

func (c *client) SetUserClientRoles(id string, clientName string, roles ...string) error {
	cl, err := c.getClient(clientName)
	if err != nil {
		return err
	}

	currentClientRoles, err := c.getUserClientRoles(id, cl.Id)
	if err != nil {
		return err
	}
	if deleteRoles := lo.Filter(currentClientRoles, func(role Role, _ int) bool { return !lo.Contains(roles, role.Name) }); len(deleteRoles) > 0 {
		if err = c.deleteUserClientRoles(
			id,
			cl.Id,
			deleteRoles,
		); err != nil {
			return err
		}
	}

	addRoles := []string{}
	for _, role := range roles {
		if !lo.ContainsBy(currentClientRoles, func(r Role) bool {
			return r.Name == role
		}) {
			addRoles = append(addRoles, role)
		}
	}

	if len(addRoles) == 0 {
		return nil
	}

	clientRoles, err := c.getClientRoles(cl.Id)
	if err != nil {
		return err
	}

	userClientRoles := lo.Filter(clientRoles, func(role Role, _ int) bool {
		return lo.Contains(addRoles, role.Name)
	})

	if err := c.addUserClientRoles(id, cl.Id, userClientRoles); err != nil {
		return err
	}

	return nil
}

func (c *client) AddUserClientRoles(id string, clientName string, roles ...string) error {
	cl, err := c.getClient(clientName)
	if err != nil {
		return err
	}

	clientRoles, err := c.getClientRoles(cl.Id)
	if err != nil {
		return err
	}

	userClientRoles := lo.Filter(clientRoles, func(role Role, _ int) bool {
		return lo.Contains(roles, role.Name)
	})

	if err := c.addUserClientRoles(id, cl.Id, userClientRoles); err != nil {
		return err
	}

	return nil
}

func (c *client) DeleteUserClientRoles(id string, clientName string, roles ...string) error {
	_, err := c.getClient(clientName)
	if err != nil {
		return err
	}
	return nil
}
