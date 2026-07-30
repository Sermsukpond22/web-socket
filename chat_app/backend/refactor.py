import re

service_file = 'modules/auth/auth.service.go'
with open(service_file, 'r', encoding='utf-8') as f:
    content = f.read()

# Add os import
if '"os"' not in content:
    content = content.replace('"strings"', '"os"\n\t"strings"')

# Add GenerateRefreshToken and ValidateRefreshToken to interface
content = re.sub(r'(Login\(input LoginInput\) \(\*models\.User, string, error\))', r'Login(input LoginInput) (*models.User, string, string, error)', content)
content = content.replace('GenerateToken(user *models.User) (string, error)', 'GenerateToken(user *models.User) (string, error)\n\tGenerateRefreshToken(user *models.User) (string, error)\n\tValidateRefreshToken(tokenString string) (*jwt.Token, jwt.MapClaims, error)')

# Update NewAuthService to fail if jwtSecret is empty, but wait, main.go already fails. We can leave NewAuthService logic but remove default fallback.
content = re.sub(r'if jwtSecret == "" \{\s*jwtSecret = "default_jwt_secret_key_chat_app"\s*\}', '', content)

# Update Login return signature
content = re.sub(r'func \(s \*authService\) Login\(input LoginInput\) \(\*models\.User, string, error\) \{', r'func (s *authService) Login(input LoginInput) (*models.User, string, string, error) {', content)
content = re.sub(r'return nil, "", errors\.New', r'return nil, "", "", errors.New', content)
content = re.sub(r'return nil, "", fmt\.Errorf', r'return nil, "", "", fmt.Errorf', content)
content = content.replace('return user, token, nil', '''
	refreshToken, err := s.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return user, token, refreshToken, nil
''')

# Add GenerateRefreshToken and ValidateRefreshToken implementations
new_methods = '''
func (s *authService) GenerateRefreshToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = s.jwtSecret // fallback
	}
	return token.SignedString([]byte(refreshSecret))
}

func (s *authService) ValidateRefreshToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = s.jwtSecret // fallback
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(refreshSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, errors.New("invalid refresh token")
}
'''
if 'func (s *authService) GenerateRefreshToken' not in content:
    content += new_methods

with open(service_file, 'w', encoding='utf-8') as f:
    f.write(content)

controller_file = 'modules/auth/auth.controller.go'
with open(controller_file, 'r', encoding='utf-8') as f:
    c_content = f.read()

# Update Login handler
c_content = re.sub(r'user, token, err := c\.authService\.Login\(input\)', r'user, token, refreshToken, err := c.authService.Login(input)', c_content)
c_content = re.sub(r'"token": token,\s*"user":  user,', r'"token": token,\n\t\t"refresh_token": refreshToken,\n\t\t"user":  user,', c_content)

# Add RefreshToken handler
refresh_handler = '''
func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
	type RefreshInput struct {
		RefreshToken string json:"refresh_token"
	}
	var input RefreshInput
	if err := ctx.BodyParser(&input); err != nil || input.RefreshToken == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, claims, err := c.authService.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	userID := uint(claims["user_id"].(float64))
	user, err := c.authService.GetUserByID(userID)
	if err != nil || user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	newToken, err := c.authService.GenerateToken(user)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": newToken,
	})
}
'''
if 'func (c *AuthController) RefreshToken' not in c_content:
    c_content += refresh_handler

with open(controller_file, 'w', encoding='utf-8') as f:
    f.write(c_content)

routes_file = 'modules/auth/auth.routes.go'
with open(routes_file, 'r', encoding='utf-8') as f:
    r_content = f.read()

if 'auth.Post("/refresh"' not in r_content:
    r_content = r_content.replace('auth.Post("/login", authController.Login)', 'auth.Post("/login", authController.Login)\\n\\tauth.Post("/refresh", authController.RefreshToken)')

with open(routes_file, 'w', encoding='utf-8') as f:
    f.write(r_content)
