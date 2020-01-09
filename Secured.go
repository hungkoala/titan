package titan

// no security roles are allowed.
func DenyAll() AuthFunc {
	return func(*Context) bool {
		return false
	}
}

func IsAuthenticated() AuthFunc {
	return func(ctx *Context) bool {
		return ctx.UserInfo() != nil
	}
}

func IsAnonymous() AuthFunc {
	return func(*Context) bool {
		return true
	}
}

func Secured(roles ...string) AuthFunc {
	return func(ctx *Context) bool {
		if len(roles) == 0 {
			return false
		}
		userInfo := ctx.UserInfo()

		if userInfo == nil {
			return false
		}

		roleStr := string(userInfo.Role)

		if roleStr == "" {
			return false
		}

		for _, r := range roles {
			if r == roleStr {
				return true
			}
		}
		return false
	}
}
