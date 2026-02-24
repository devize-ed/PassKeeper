# Package service

Business logic layer for PassKeeper.
AuthService and ItemService delegate to Storage; ItemService reads user ID from context (set by auth interceptor).
