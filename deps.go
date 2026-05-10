//temporary file to install dependencies

package main

import (
    _ "github.com/labstack/echo/v4"
    _ "github.com/stretchr/testify"
    _ "gorm.io/driver/postgres"
    _ "gorm.io/driver/sqlite"
    _ "gorm.io/gorm"
)