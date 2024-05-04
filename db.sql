CREATE TABLE car (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Identificador único, generado automáticamente
    online BOOLEAN NOT NULL, -- Estado en línea o fuera de línea
    brand TEXT NOT NULL, -- Marca del automóvil
    model TEXT NOT NULL, -- Modelo del automóvil
    year INTEGER NOT NULL, -- Año de fabricación
    kilometers BIGINT NOT NULL, -- Kilometraje
    car_domain TEXT NOT NULL, -- Dominio del automóvil
    price NUMERIC(15, 2) NOT NULL, -- Precio del automóvil, con precisión de 15 dígitos y 2 decimales
    info_price NUMERIC(15, 2) NOT NULL, -- Información sobre el precio, con precisión de 15 dígitos y 2 decimales
    currency TEXT NOT NULL SET DEFAULT 'USD'-- Moneda en la que se muestra el precio (por ejemplo, USD o ARS)
);


