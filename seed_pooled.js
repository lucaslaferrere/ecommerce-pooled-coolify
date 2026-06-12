// ============================================================
// SEED SCRIPT - POOLED ECOMMERCE
// Correr con:
// mongosh "mongodb://root:PASSWORD@HOST:27017/?directConnection=true" seed_pooled.js
// ============================================================

const db = connect("mongodb://root:Q62XCH20VplJvaCzy6VM0I53mIzmrZs4wbwNpfJBwCcOyEix7Up7AzEt7zmw9pSH@m48koc8ckkkkg088c84o8880:27017/ecommerce?directConnection=true&authSource=admin");

const now = new Date();

// ============================================================
// USUARIOS
// ============================================================
db.users.deleteMany({});

db.users.insertOne({
  email: "admin@pooled.com.ar",
  password_hash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password123
  role: "admin",
  created_at: now,
  updated_at: now
});

print("✅ Usuario admin insertado");

// ============================================================
// PRODUCTOS
// ============================================================
db.products.deleteMany({});

const products = [

  // ─── LUMINARIAS ───────────────────────────────────────────

  {
    name: "OSIRE RGBW 18W",
    description: "Luminaria subacuática 24VDC con tecnología RGBW de 4 canales independientes. Cuerpo en acero inoxidable macizo mecanizado con lente de vidrio y difusor integrado. Protección térmica activa y alcance de señal de 300 metros.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Voltaje de entrada", value: "DC 24V" },
      { key: "Potencia máxima total", value: "18W" },
      { key: "Potencia por canal", value: "9W — Rojo / Verde / Azul / Blanco" },
      { key: "Tipo de luz", value: "RGBW (4 canales independientes)" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +60°C" },
      { key: "Protección térmica", value: "Activa — atenúa automáticamente al superar umbral térmico" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Vida útil estimada", value: "50.000 horas" },
      { key: "Conexión", value: "3 hilos: Rojo (V+) · Negro (V-) · Amarillo (DATOS)" },
      { key: "Longitud máxima de señal", value: "300 metros" },
      { key: "Cuerpo", value: "Acero inoxidable macizo mecanizado" },
      { key: "Lente", value: "Vidrio con difusor integrado" },
      { key: "Dimensiones del frente", value: "100 x 18mm" },
      { key: "Garantía", value: "3 años" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "HORUS RGBW 18W",
    description: "Luminaria subacuática 12VDC de perfil ultraplano (8mm) con tecnología RGBW de 5 hilos. Alta potencia de proyección con 18W totales, ideal para piscinas de gran escala o profundidad.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "18W" },
      { key: "Potencia por canal", value: "7,2W — Rojo / Verde / Azul / Blanco" },
      { key: "Tipo de luz", value: "RGBW (4 canales independientes)" },
      { key: "Lumens", value: "110 lm/W" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +30°C" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Conexión", value: "5 hilos: Blanco (V+) · Rojo (R-) · Verde (G-) · Azul (B-) · Amarillo (W-)" },
      { key: "Dimensiones del frente", value: "100 x 8mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "NAZAR RGBW 9W",
    description: "Luminaria subacuática 12VDC de perfil ultraplano (8mm) con tecnología RGBW. Optimizada para instalaciones de acento, balizamiento y señalización. Permite mayor cantidad de unidades por fuente.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "9W" },
      { key: "Potencia por canal", value: "3,6W — Rojo / Verde / Azul / Blanco" },
      { key: "Tipo de luz", value: "RGBW (4 canales independientes)" },
      { key: "Lumens", value: "110 lm/W" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +30°C" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Conexión", value: "5 hilos: Blanco (V+) · Rojo (R-) · Verde (G-) · Azul (B-) · Amarillo (W-)" },
      { key: "Dimensiones del frente", value: "80 x 8mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "HORUS W&W 7.2W",
    description: "Luminaria subacuática 12VDC de perfil ultraplano con tecnología CCT (temperatura de color ajustable). Exclusiva para proyectos que requieren luz blanca de alta calidad con cableado simplificado de 3 hilos.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "7,2W" },
      { key: "Potencia por canal", value: "7,2W — Blanco frío (CW) / Blanco cálido (WW)" },
      { key: "Tipo de luz", value: "CCT — Temperatura de color ajustable" },
      { key: "Lumens", value: "120 lm/W" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +30°C" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Conexión", value: "3 hilos: Blanco (V+) · Azul (CW-) · Amarillo (WW-)" },
      { key: "Dimensiones del frente", value: "100 x 8mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT RGBW CHAPA APLICAR 15W",
    description: "Luminaria subacuática de superficie en acero inoxidable con óptica de 120°. Instalación rápida y no invasiva, ideal para piscinas ya construidas o de fibra de vidrio.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-CHAPA-RGBW-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "15W" },
      { key: "Potencia por canal", value: "6W — Rojo / Verde / Azul / Blanco" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de luz", value: "RGBW (4 canales independientes)" },
      { key: "Lumens", value: "75 lm/W" },
      { key: "Material", value: "INOX" },
      { key: "Dimensiones del frente", value: "145 x 24mm" },
      { key: "Garantía", value: "6 meses" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT APLICAR RGBW 6.25W",
    description: "Luminaria subacuática de superficie con óptica de 120° y eficiencia de 100 lm/W. Disponible en materiales PC e INOX. Instalación sin mampostería, ideal para remodelaciones.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-APLICAR-RGBW-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-APLICAR-RGBW-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "6,25W" },
      { key: "Potencia por canal", value: "2,5W — Rojo / Verde / Azul / Blanco" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de luz", value: "RGBW (4 canales independientes)" },
      { key: "Lumens", value: "100 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo)" },
      { key: "Dimensiones del frente", value: "120 x 18mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT APLICAR WHITE 10W",
    description: "Luminaria subacuática de superficie con luz blanca fría de alta potencia. Proyecta un haz denso para resaltar la cristalinidad del agua. Disponible en PC e INOX.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-APLICAR-WHITE-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-APLICAR-WHITE-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "10W" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de luz", value: "Blanco frío" },
      { key: "Lumens", value: "100 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo)" },
      { key: "Dimensiones del frente", value: "120 x 18mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT EMPOTRAR RGBW 9.4W",
    description: "Luminaria subacuática empotrada con acabado a ras de pared mediante aro embellecedor. La solución de mayor integración arquitectónica para piscinas de hormigón en construcción.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-EMPOTRAR-RGBW-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-EMPOTRAR-RGBW-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "9,4W" },
      { key: "Potencia por canal", value: "3,75W — Rojo / Verde / Azul / Blanco" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de luz", value: "RGBW (4 canales independientes)" },
      { key: "Lumens", value: "75 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo de aro embellecedor)" },
      { key: "Dimensiones del frente", value: "100 x 17mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT EMPOTRAR WHITE 15W",
    description: "Luminaria subacuática empotrada con luz blanca fría de máxima potencia. Integración arquitectónica perfecta con acabado a ras de pared para piscinas de hormigón.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-EMPOTRAR-WHITE-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-EMPOTRAR-WHITE-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],
    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia máxima total", value: "15W" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de luz", value: "Blanco frío" },
      { key: "Lumens", value: "75 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo de aro embellecedor)" },
      { key: "Dimensiones del frente", value: "100 x 17mm" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  // ─── CONTROLADORES ────────────────────────────────────────

  {
    name: "TE3 RGBW 60W",
    description: "Controlador inteligente con fuente de 60W para proyectos compactos. Conectividad triple WiFi + Bluetooth + RF 2.4GHz con integración al ecosistema Tuya Smart (Alexa / Google Home). Clase II, sin toma a tierra requerida.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador RGBW" },
      { key: "Entrada fuente", value: "90 ~ 264VAC (soporta picos de 300VAC por 5 seg)" },
      { key: "Potencia máxima", value: "60W" },
      { key: "Clase eléctrica", value: "Clase II (sin toma a tierra requerida)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Carcasa", value: "Plástica encapsulada" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema smart home", value: "Tuya Smart — Amazon Alexa / Google Home" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE6 RGBW 102W",
    description: "Controlador inteligente con fuente de 102W / 8,5A. El estándar residencial para piscinas domiciliarias. Conectividad triple WiFi + Bluetooth + RF 2.4GHz con integración Tuya Smart.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador RGBW" },
      { key: "Entrada fuente", value: "90 ~ 264VAC (soporta picos de 300VAC por 5 seg)" },
      { key: "Potencia máxima", value: "102W" },
      { key: "Corriente nominal", value: "8,5A" },
      { key: "Clase eléctrica", value: "Clase II (sin toma a tierra requerida)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Carcasa", value: "Plástica encapsulada" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema smart home", value: "Tuya Smart — Amazon Alexa / Google Home" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE9 RGBW 150W",
    description: "Controlador industrial con fuente de 150W / 12,5A para proyectos de gran formato. Chasis metálico Clase I con protección Surge 6KV/4KV y rango de entrada ultra amplio 100~305VAC.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador RGBW" },
      { key: "Entrada fuente", value: "100 ~ 305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC · 12,5A · 150W" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Clase eléctrica", value: "Clase I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema smart home", value: "Tuya Smart — Amazon Alexa / Google Home" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE15 RGBW 300W Dual",
    description: "Controlador industrial de doble salida independiente (2x150W = 300W totales). Permite alimentar dos circuitos desde un único gabinete. Ideal para proyectos de zonificación como piscina + spa.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador RGBW dual" },
      { key: "Entrada fuente", value: "100 ~ 305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC x2 · 12,5A x2 · 300W totales" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Clase eléctrica", value: "Clase I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema smart home", value: "Tuya Smart — Amazon Alexa / Google Home" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE3 W&W 60W",
    description: "Controlador CCT con fuente de 60W para proyectos pequeños de iluminación blanca ajustable. Control RF sin WiFi ni Bluetooth, con dimming de 20 niveles y memoria de estado.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador CCT (Blanco ajustable)" },
      { key: "Entrada fuente", value: "90 ~ 264VAC (soporta picos de 300VAC por 5 seg)" },
      { key: "Salida", value: "12VDC · 5A · 60W" },
      { key: "Clase eléctrica", value: "Clase II (sin toma a tierra)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles de precisión" },
      { key: "Memoria", value: "Resume function — recupera parámetros previos al corte" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE6 W&W 102W",
    description: "Controlador CCT con fuente de 102W / 8,5A. Estándar residencial para piscinas con iluminación blanca ajustable. Control RF con dimming de 20 niveles y memoria de estado.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador CCT (Blanco ajustable)" },
      { key: "Entrada fuente", value: "90 ~ 264VAC (soporta picos de 300VAC por 5 seg)" },
      { key: "Salida", value: "12VDC · 8,5A · 102W" },
      { key: "Clase eléctrica", value: "Clase II (sin toma a tierra)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles de precisión" },
      { key: "Memoria", value: "Resume function — recupera parámetros previos al corte" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE9 W&W 150W",
    description: "Controlador CCT industrial con fuente de 150W / 12,5A. Chasis metálico Clase I con Surge 6KV/4KV y PFC activo. Para hoteles, clubes y espejos de agua de gran envergadura.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador CCT (Blanco ajustable)" },
      { key: "Entrada fuente", value: "100 ~ 305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC · 12,5A · 150W" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Clase eléctrica", value: "Clase I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles de precisión" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "TE15 W&W 300W Dual",
    description: "Controlador CCT de doble salida independiente (2x150W = 300W totales). Ideal para proyectos de zonificación con luz blanca ajustable. Chasis metálico industrial con Surge 6KV/4KV.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador CCT dual (Blanco ajustable)" },
      { key: "Entrada fuente", value: "100 ~ 305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC x2 · 12,5A x2 · 300W totales" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Clase eléctrica", value: "Clase I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles de precisión" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "PCE-100-SYS 24V 96W",
    description: "Controlador SYS con fuente 24VDC para uso exclusivo con luminarias OSIRE. Control RF 2.4GHz con auto-transmisión en cascada para cobertura ilimitada. Requiere Gateway separado para control por app.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador SYS 24V" },
      { key: "Entrada fuente", value: "100 ~ 305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "92%" },
      { key: "Salida", value: "24VDC · 4A · 96W" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Clase eléctrica", value: "Clase I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Transmisión", value: "RF 2.4GHz — alcance hasta 30 metros" },
      { key: "Control por app", value: "Sí (requiere Gateway 2.4GHz separado)" },
      { key: "Auto-transmisión cascada", value: "Sí — cobertura ilimitada" },
      { key: "Compatibilidad", value: "Exclusivo para luminarias 24V (OSIRE)" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "PCE-60-12 Controlador 5 en 1",
    description: "Controlador versátil 5 en 1 compatible con modos Dimmer, CCT, RGB, RGBW y RGB+CCT. Conectividad triple WiFi + Bluetooth + RF 2.4GHz con auto-transmisión en cascada. Compatible con tiras LED 12VDC comerciales.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Tipo", value: "Controlador 5 en 1 (Dimmer / CCT / RGB / RGBW / RGB+CCT)" },
      { key: "Entrada fuente", value: "100 ~ 240VAC · 47~63Hz" },
      { key: "Salida", value: "12VDC · 5A · 60W" },
      { key: "Protección IP fuente", value: "IP67" },
      { key: "Clase eléctrica", value: "Clase I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Plástica encapsulada" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema smart home", value: "Tuya Smart — Amazon Alexa / Google Home" },
      { key: "Auto-transmisión cascada", value: "Sí — cobertura ilimitada" },
      { key: "Garantía", value: "1 año" }
    ],
    created_at: now,
    updated_at: now
  },

  // ─── PRÓXIMAMENTE ─────────────────────────────────────────

  {
    name: "OSIRE RGB+CCT 18W (Próximamente)",
    description: "Próxima luminaria subacuática 24VDC con tecnología RGB+CCT de 5 canales independientes. Combina escenas dinámicas en color con luz blanca real en tonalidades frías o cálidas. Cuerpo en acero inoxidable con lente de vidrio.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Voltaje de entrada", value: "DC 24V" },
      { key: "Potencia máxima total", value: "18W" },
      { key: "Potencia por canal", value: "9W — R/G/B · 18W Blanco frío / 18W Blanco cálido" },
      { key: "Tipo de luz", value: "RGB+CCT (5 canales independientes)" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +60°C" },
      { key: "Protección térmica", value: "Activa" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Vida útil estimada", value: "50.000 horas" },
      { key: "Conexión", value: "3 hilos: Rojo (V+) · Negro (V-) · Amarillo (DATOS)" },
      { key: "Longitud máxima de señal", value: "300 metros" },
      { key: "Cuerpo", value: "Acero inoxidable macizo mecanizado" },
      { key: "Garantía", value: "3 años" },
      { key: "Disponibilidad", value: "Próximamente" }
    ],
    created_at: now,
    updated_at: now
  },

  {
    name: "OSIRE RGB+CCT 9W (Próximamente)",
    description: "Próxima luminaria subacuática 24VDC con tecnología RGB+CCT de 5 canales independientes y 9W. Versión compacta de la familia OSIRE con las mismas prestaciones de alto rendimiento.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],
    specs: [
      { key: "Voltaje de entrada", value: "DC 24V" },
      { key: "Potencia máxima total", value: "9W" },
      { key: "Potencia por canal", value: "4,5W — R/G/B · 9W Blanco frío / 9W Blanco cálido" },
      { key: "Tipo de luz", value: "RGB+CCT (5 canales independientes)" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +60°C" },
      { key: "Protección térmica", value: "Activa" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Vida útil estimada", value: "50.000 horas" },
      { key: "Conexión", value: "3 hilos: Rojo (V+) · Negro (V-) · Amarillo (DATOS)" },
      { key: "Longitud máxima de señal", value: "300 metros" },
      { key: "Cuerpo", value: "Acero inoxidable macizo mecanizado" },
      { key: "Garantía", value: "3 años" },
      { key: "Disponibilidad", value: "Próximamente" }
    ],
    created_at: now,
    updated_at: now
  }

];

db.products.insertMany(products);
print(`✅ ${products.length} productos insertados`);

// ============================================================
// RESUMEN
// ============================================================
print("\n📊 RESUMEN:");
print(`   users:    ${db.users.countDocuments()}`);
print(`   products: ${db.products.countDocuments()}`);
print(`   kits:     ${db.kits.countDocuments()}`);
print("\n✅ Seed completado exitosamente");
