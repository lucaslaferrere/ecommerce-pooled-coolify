// ============================================================
// SEED SCRIPT v2 - POOLED ECOMMERCE
// Basado en: Fichas Técnicas WEB (doc actualizado)
// Correr con: mongosh "MONGO_URI" seed_pooled_v2.js
// ============================================================

const db = connect("mongodb://root:Q62XCH20VplJvaCzy6VM0I53mIzmrZs4wbwNpfJBwCcOyEix7Up7AzEt7zmw9pSH@m48koc8ckkkkg088c84o8880:27017/ecommerce?directConnection=true&authSource=admin");

const now = new Date();

// ============================================================
// USUARIOS
// ============================================================
db.users.deleteMany({});
db.users.insertOne({
  email: "admin@pooled.com.ar",
  password_hash: "$2b$10$g3O81giXzGFknksw5rkQ..veEbtN7iWyijsbNSte9fMC8EhRMp6HS", // Pooled2026!xK9m
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

  // ─── LUMINARIAS ─────────────────────────────────────────

  {
    name: "OSIRE 18W RGBW",
    subtitle: "Acero inoxidable | 3 años de garantía | Tecnología RGBW profesional",
    description: "Luminaria premium subacuática 24VDC con tecnología RGBW de 4 canales independientes. Cuerpo en acero inoxidable 316L macizo mecanizado con lente de vidrio templado y difusor integrado. Protección térmica activa y alcance de señal de 300 metros.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Seguridad Total", description: "Funciona con 24V de baja tensión. Cero riesgo de descargas eléctricas. 100% seguro para vos y tu familia." },
      { title: "Colores Vibrantes + Blanco Puro", description: "Tecnología RGBW: 4 canales independientes (Rojo, Verde, Azul y Blanco dedicado). Pasá de fiesta colorida a elegancia minimalista sin tonos azulados." },
      { title: "Construcción Premium", description: "Cuerpo de acero inoxidable macizo mecanizado. Lente de vidrio templado. Diseñado para durar décadas." },
      { title: "Inteligencia Térmica", description: "Se autoprotege del calor: si detecta temperatura alta, reduce brillo automáticamente. No se quema ni se rompe." },
      { title: "Resistencia Extrema", description: "Sumergible continuo (IP68). Funciona desde -20°C hasta +60°C. Resiste cloro, sal y productos químicos." }
    ],

    main_specs: [
      { key: "Potencia", value: "18W", meaning: "Consume muy poco (como 1 lámpara LED hogareña)" },
      { key: "Brillo", value: "160 lúmenes", meaning: "Luz potente y clara para piletas medianas/grandes" },
      { key: "Tipo de luz", value: "RGBW (4 canales)", meaning: "Colores + blanco puro real" },
      { key: "Voltaje", value: "24VDC", meaning: "Baja tensión = seguridad total" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente, resistente al polvo" },
      { key: "Temperatura", value: "-20°C a +60°C", meaning: "Funciona en cualquier clima" },
      { key: "Garantía", value: "3 años", meaning: "La más larga del mercado" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 24V" },
      { key: "Potencia Máxima Total", value: "18W" },
      { key: "Potencia por Canal", value: "9W (Rojo / Verde / Azul / Blanco)" },
      { key: "Tipo de Luz", value: "RGBW (4 canales independientes)" },
      { key: "Flujo luminoso", value: "160 lúmenes" },
      { key: "Material del cuerpo", value: "Acero inoxidable 316L macizo mecanizado" },
      { key: "Material del lente", value: "Vidrio templado con difusor integrado" },
      { key: "Dimensiones del frente", value: "100 × 18mm" },
      { key: "Sistema de conexión", value: "3 hilos: Rojo (V+) · Negro (V-) · Amarillo (DATOS)" },
      { key: "Longitud máxima de señal", value: "300 metros" },
      { key: "Protocolo", value: "Digital 3 hilos" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +60°C" },
      { key: "Protección Térmica Activa", value: "Sí (atenúa automáticamente si supera umbral térmico)" },
      { key: "Resistencia química", value: "Cloro, sal, pH ácido/alcalino" },
      { key: "Vida útil estimada", value: "50.000 horas (L70)" },
      { key: "Garantía", value: "3 años" },
      { key: "Controladores compatibles", value: "Serie PCE-100-SYS (24V)" },
      { key: "Instalación", value: "Embutir / Aplicar (según versión)" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "HORUS 18W RGBW",
    subtitle: "Perfil ultrachato 8mm | 1 año de garantía | Proyección superior",
    description: "Luminaria subacuática 12VDC de perfil ultraplano (8mm) con tecnología RGBW de 5 hilos. Alta potencia de proyección con 18W totales, ideal para piscinas de gran escala o profundidad.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Diseño Ultraplano Innovador", description: "Solo 8mm de espesor. Casi invisible en la pared de la pileta. Estética minimalista y moderna." },
      { title: "Máxima Seguridad Subacuática", description: "Funciona con 12V de extra baja tensión. Cero riesgo de electrocución. Certificado para uso en piletas." },
      { title: "Doble Potencia de Proyección", description: "18W totales (el doble que el NAZAR). Ideal para piletas grandes o profundas. Más luz, más cobertura." },
      { title: "Colores + Blanco Real (RGBW)", description: "4 canales independientes: transicionás de fiesta colorida a iluminación elegante blanca sin tonos azulados." },
      { title: "Resistencia Total IP68", description: "Sumergible permanente. Resiste cloro, sal y productos químicos. Protección absoluta." }
    ],

    main_specs: [
      { key: "Potencia", value: "18W total", meaning: "Alta potencia para piletas grandes" },
      { key: "Potencia por canal", value: "7,2W", meaning: "Doble intensidad vs. NAZAR (más brillo)" },
      { key: "Eficiencia", value: "110 lúmenes/watt", meaning: "Muy eficiente (ahorro energético)" },
      { key: "Espesor", value: "8mm", meaning: "Perfil ultrachato, casi invisible" },
      { key: "Tipo de luz", value: "RGBW (4 canales)", meaning: "Colores + blanco puro" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Temperatura", value: "-20°C a +30°C", meaning: "Funciona en la mayoría de climas" },
      { key: "Dimensiones", value: "100mm × 8mm", meaning: "Compacto y elegante" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "18W" },
      { key: "Potencia por Canal", value: "7,2W (Rojo / Verde / Azul / Blanco)" },
      { key: "Tipo de Luz", value: "RGBW (4 canales independientes)" },
      { key: "Eficiencia luminosa", value: "110 lm/W" },
      { key: "Sistema de conexión", value: "5 hilos: Blanco (V+) · Rojo (R-) · Verde (G-) · Azul (B-) · Amarillo (W-)" },
      { key: "Dimensiones del frente", value: "100mm × 8mm" },
      { key: "Perfil", value: "Ultrachato (8mm de espesor total)" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +30°C" },
      { key: "Garantía", value: "1 año" },
      { key: "Controladores compatibles", value: "Serie TE RGBW (12V)" },
      { key: "Instalación", value: "Empotrar" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "NAZAR 9W RGBW",
    subtitle: "Perfil ultrachato 8mm | 1 año de garantía | Ideal para acento y balizamiento",
    description: "Luminaria subacuática 12VDC de perfil ultraplano (8mm) con tecnología RGBW. Optimizada para instalaciones de acento, balizamiento y señalización. Permite mayor cantidad de unidades por fuente.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Diseño Ultraplano Premium", description: "Solo 8mm de espesor. Integración casi invisible. Estética limpia y minimalista." },
      { title: "100% Seguro para Piletas", description: "Funciona con 12V extra baja tensión. Eliminás cualquier riesgo eléctrico. Certificado para inmersión continua." },
      { title: "Eficiencia Energética Superior", description: "Con 9W consume la mitad que el HORUS. Podés conectar el doble de luces al mismo controlador. Menor costo operativo." },
      { title: "Colores Vibrantes + Blanco Puro", description: "Tecnología RGBW con 4 canales independientes. Pasá de ambientes festivos a iluminación arquitectónica elegante." },
      { title: "Perfecto para Balizamiento", description: "Ideal para escaleras sumergidas, bordes, bancos de hidromasaje. Marca el perímetro sin deslumbrar." }
    ],

    main_specs: [
      { key: "Potencia", value: "9W total", meaning: "Consumo eficiente (mitad del HORUS)" },
      { key: "Potencia por canal", value: "3,6W", meaning: "Optimizado para acento y balizamiento" },
      { key: "Eficiencia", value: "110 lúmenes/watt", meaning: "Máximo ahorro energético" },
      { key: "Espesor", value: "8mm", meaning: "Perfil ultrachato, mínima intrusión" },
      { key: "Diámetro", value: "80mm", meaning: "Más compacto que HORUS (100mm)" },
      { key: "Tipo de luz", value: "RGBW (4 canales)", meaning: "Colores + blanco real" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "9W" },
      { key: "Potencia por Canal", value: "3,6W (Rojo / Verde / Azul / Blanco)" },
      { key: "Tipo de Luz", value: "RGBW (4 canales independientes)" },
      { key: "Eficiencia luminosa", value: "110 lm/W" },
      { key: "Sistema de conexión", value: "5 hilos: Blanco (V+) · Rojo (R-) · Verde (G-) · Azul (B-) · Amarillo (W-)" },
      { key: "Dimensiones del frente", value: "80mm × 8mm" },
      { key: "Perfil", value: "Ultrachato (8mm de espesor total)" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +30°C" },
      { key: "Garantía", value: "1 año" },
      { key: "Controladores compatibles", value: "Serie TE RGBW (12V)" },
      { key: "Instalación", value: "Empotrar" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "HORUS W&W 7.2W",
    subtitle: "Perfil ultrachato 8mm | 1 año de garantía | Temperatura de color regulable",
    description: "Luminaria subacuática 12VDC de perfil ultraplano con tecnología CCT (temperatura de color ajustable). Exclusiva para proyectos que requieren luz blanca de alta calidad con cableado simplificado de 3 hilos.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Luz Blanca de Alta Calidad", description: "Exclusivamente luz blanca. Sin colores. Ideal para piletas que buscan elegancia y pureza todo el tiempo." },
      { title: "Ajustá el Tono de Blanco", description: "Cambiá entre blanco cálido (amarillento, relajante) y blanco frío (brillante, energizante). Todo con el mismo equipo." },
      { title: "Diseño Ultraplano 8mm", description: "Integración arquitectónica limpia. Casi no se nota en la pared. Estética minimalista premium." },
      { title: "Instalación Simplificada", description: "Solo 3 cables vs. 5 cables del RGBW. Instalación más rápida. Menos puntos de falla. Más confiable." },
      { title: "100% Seguro", description: "Funciona con 12V de extra baja tensión. Cero riesgo de descargas. Certificado para inmersión continua." }
    ],

    main_specs: [
      { key: "Potencia", value: "7,2W", meaning: "Consumo eficiente" },
      { key: "Eficiencia", value: "120 lúmenes/watt", meaning: "Máxima eficiencia de la línea" },
      { key: "Espesor", value: "8mm", meaning: "Perfil ultrachato" },
      { key: "Tipo de luz", value: "CCT (Blanco ajustable)", meaning: "De cálido a frío, sin colores" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Cableado", value: "3 hilos", meaning: "Instalación simplificada" },
      { key: "Dimensiones", value: "100mm × 8mm", meaning: "Compacto y elegante" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "7,2W" },
      { key: "Tipo de Luz", value: "CCT (Temperatura de Color Ajustable)" },
      { key: "Eficiencia luminosa", value: "120 lm/W" },
      { key: "Sistema de conexión", value: "3 hilos: Blanco (V+) · Azul (CW-) · Amarillo (WW-)" },
      { key: "Dimensiones del frente", value: "100mm × 8mm" },
      { key: "Perfil", value: "Ultrachato (8mm de espesor total)" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Temperatura de trabajo", value: "-20°C ~ +30°C" },
      { key: "Garantía", value: "1 año" },
      { key: "Controladores compatibles", value: "Serie TE W&W (12V)" },
      { key: "Instalación", value: "Empotrar" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT Aplicar RGBW 6.25W",
    subtitle: "Sin obra | 1 año de garantía | Instalación rápida",
    description: "Luminaria subacuática de superficie con óptica de 120° y eficiencia de 100 lm/W. Disponible en materiales PC e INOX. Instalación sin mampostería, ideal para remodelaciones y piletas existentes.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-APLICAR-RGBW-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-APLICAR-RGBW-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],

    benefits: [
      { title: "Sin Obra", description: "Se aplica sobre la superficie de la pileta. No necesitás romper ni embutir. Instalación rápida y no invasiva." },
      { title: "Colores + Blanco Puro RGBW", description: "4 canales independientes. Colores festivos o blanco elegante sin tonos azulados." },
      { title: "100% Seguro", description: "12V de baja tensión. Sumergible permanente IP68." }
    ],

    main_specs: [
      { key: "Potencia", value: "6,25W", meaning: "Consumo muy eficiente" },
      { key: "Eficiencia", value: "100 lúmenes/watt", meaning: "Excelente rendimiento" },
      { key: "Ángulo", value: "120°", meaning: "Amplia cobertura lumínica" },
      { key: "Tipo de luz", value: "RGBW (4 canales)", meaning: "Colores + blanco puro" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Dimensiones", value: "120 × 18mm", meaning: "Compacto" },
      { key: "Material", value: "PC / INOX", meaning: "Elegí según tu preferencia" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "6,25W" },
      { key: "Potencia por Canal", value: "2,5W (Rojo / Verde / Azul / Blanco)" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de Luz", value: "RGBW (4 canales independientes)" },
      { key: "Eficiencia luminosa", value: "100 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo)" },
      { key: "Dimensiones del frente", value: "120 × 18mm" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT Aplicar WHITE 10W",
    subtitle: "Sin obra | 1 año de garantía | Blanco puro de alta potencia",
    description: "Luminaria subacuática de superficie con luz blanca fría de alta potencia. Proyecta un haz denso para resaltar la cristalinidad del agua. Disponible en PC e INOX.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-APLICAR-WHITE-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-APLICAR-WHITE-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],

    benefits: [
      { title: "Sin Obra", description: "Instalación sobre superficie sin romper ni embutir. Ideal para piletas existentes." },
      { title: "Blanco Puro de Alta Potencia", description: "10W de luz blanca fría. Resalta la cristalinidad del agua." },
      { title: "100% Seguro", description: "12V de baja tensión. Sumergible permanente IP68." }
    ],

    main_specs: [
      { key: "Potencia", value: "10W", meaning: "Alta potencia en blanco puro" },
      { key: "Eficiencia", value: "100 lúmenes/watt", meaning: "Excelente rendimiento" },
      { key: "Ángulo", value: "120°", meaning: "Amplia cobertura" },
      { key: "Tipo de luz", value: "Blanco frío", meaning: "Claridad y elegancia" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "10W" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de Luz", value: "Blanco frío" },
      { key: "Eficiencia luminosa", value: "100 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo)" },
      { key: "Dimensiones del frente", value: "120 × 18mm" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT Empotrar RGBW 9.4W",
    subtitle: "Obra nueva | 1 año de garantía | Acabado a ras de pared",
    description: "Luminaria subacuática empotrada con acabado a ras de pared mediante aro embellecedor. La solución de mayor integración arquitectónica para piscinas de hormigón en construcción.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-EMPOTRAR-RGBW-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-EMPOTRAR-RGBW-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],

    benefits: [
      { title: "Acabado Arquitectónico Perfecto", description: "Embutido a ras de pared con aro embellecedor. La integración visual más limpia posible." },
      { title: "Colores + Blanco Puro RGBW", description: "4 canales independientes. Del festejo a la elegancia con un toque." },
      { title: "100% Seguro", description: "12V de baja tensión. Sumergible permanente IP68." }
    ],

    main_specs: [
      { key: "Potencia", value: "9,4W", meaning: "Consumo eficiente" },
      { key: "Eficiencia", value: "75 lúmenes/watt", meaning: "Buen rendimiento" },
      { key: "Ángulo", value: "120°", meaning: "Amplia cobertura" },
      { key: "Tipo de luz", value: "RGBW (4 canales)", meaning: "Colores + blanco puro" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Dimensiones", value: "100 × 17mm", meaning: "Compacto" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "9,4W" },
      { key: "Potencia por Canal", value: "3,75W (Rojo / Verde / Azul / Blanco)" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de Luz", value: "RGBW (4 canales independientes)" },
      { key: "Eficiencia luminosa", value: "75 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo de aro embellecedor)" },
      { key: "Dimensiones del frente", value: "100 × 17mm" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Garantía", value: "1 año" },
      { key: "Instalación", value: "Empotrar (obra nueva)" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "POOLIGHT Empotrar WHITE 15W",
    subtitle: "Obra nueva | 1 año de garantía | Blanco puro máxima potencia",
    description: "Luminaria subacuática empotrada con luz blanca fría de máxima potencia. Integración arquitectónica perfecta con acabado a ras de pared para piscinas de hormigón.",
    base_price: 0.0,
    category: "Luminarias",
    brand: "Pooled",
    images: [],
    variants: [
      { sku: "POOLIGHT-EMPOTRAR-WHITE-PC", color: "PC", size: "", stock: 0, price_adjustment: 0.0 },
      { sku: "POOLIGHT-EMPOTRAR-WHITE-INOX", color: "INOX", size: "", stock: 0, price_adjustment: 0.0 }
    ],

    benefits: [
      { title: "Acabado Arquitectónico Perfecto", description: "Embutido a ras de pared. La integración visual más limpia posible." },
      { title: "Máxima Potencia en Blanco", description: "15W de luz blanca pura. El modelo más potente de la familia POOLIGHT." },
      { title: "100% Seguro", description: "12V de baja tensión. Sumergible permanente IP68." }
    ],

    main_specs: [
      { key: "Potencia", value: "15W", meaning: "Máxima potencia en blanco" },
      { key: "Eficiencia", value: "75 lúmenes/watt", meaning: "Buen rendimiento" },
      { key: "Tipo de luz", value: "Blanco frío", meaning: "Claridad total" },
      { key: "Voltaje", value: "12VDC", meaning: "Seguridad absoluta" },
      { key: "Protección", value: "IP68", meaning: "Sumergible permanente" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Voltaje de entrada", value: "DC 12V" },
      { key: "Potencia Máxima Total", value: "15W" },
      { key: "Ángulo de apertura lumínica", value: "120°" },
      { key: "Tipo de Luz", value: "Blanco frío" },
      { key: "Eficiencia luminosa", value: "75 lm/W" },
      { key: "Material", value: "PC / INOX (según modelo de aro embellecedor)" },
      { key: "Dimensiones del frente", value: "100 × 17mm" },
      { key: "Grado de protección", value: "IP68" },
      { key: "Garantía", value: "1 año" },
      { key: "Instalación", value: "Empotrar (obra nueva)" }
    ],

    created_at: now,
    updated_at: now
  },

  // ─── CONTROLADORES ────────────────────────────────────────

  {
    name: "TE3 RGBW 60W",
    subtitle: "60W | Control por app y voz | WiFi + Bluetooth + RF | 1 año garantía",
    description: "Controlador inteligente con fuente de 60W para proyectos compactos. Conectividad triple WiFi + Bluetooth + RF 2.4GHz con integración al ecosistema Tuya Smart (Alexa / Google Home). Clase II, sin toma a tierra requerida.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Controlalo desde Tu Celular", description: "App Tuya Smart en iOS y Android. Cambiá colores, intensidad y efectos desde cualquier lugar." },
      { title: "Control por Voz", description: "Compatible con Amazon Alexa y Google Home. Decile a tu asistente y listo." },
      { title: "Triple Conectividad", description: "WiFi + Bluetooth + RF. Si se cae internet, seguís controlando por Bluetooth. Siempre funciona." },
      { title: "Expandí Sin Límites", description: "Conectá múltiples controladores en red. La señal salta de uno a otro hasta 30 metros. Cobertura ilimitada." },
      { title: "Sincronización Perfecta", description: "Todos los controladores se sincronizan automáticamente. Mismo color y velocidad en toda la pileta." },
      { title: "Protección Total", description: "Fuente IP67 resistente a humedad. Protección contra cortocircuito, sobrecarga y sobrevoltaje." }
    ],

    main_specs: [
      { key: "Potencia", value: "60W", meaning: "Para piletas pequeñas/medianas" },
      { key: "Voltaje salida", value: "12VDC", meaning: "Compatible con luminarias 12V" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF", meaning: "Triple forma de control" },
      { key: "App", value: "Tuya Smart", meaning: "Control desde celular" },
      { key: "Voz", value: "Alexa + Google Home", meaning: "Comandos de voz" },
      { key: "Protección fuente", value: "IP67", meaning: "Resistente a humedad" },
      { key: "Rango voltaje", value: "90-264VAC", meaning: "Funciona con variaciones de red" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "90~264VAC (rango universal)" },
      { key: "Tolerancia picos", value: "Hasta 300VAC por 5 segundos" },
      { key: "Salida", value: "12VDC · 5A · 60W máximo" },
      { key: "Clase", value: "II (sin toma a tierra requerida)" },
      { key: "Carcasa", value: "Plástica encapsulada" },
      { key: "Protección IP", value: "IP67" },
      { key: "Protecciones activas", value: "Cortocircuito / Sobrecarga / Sobrevoltaje" },
      { key: "Conectividad", value: "WiFi 2.4GHz + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema", value: "Tuya Smart" },
      { key: "Compatibilidad voz", value: "Amazon Alexa / Google Home" },
      { key: "Alcance RF", value: "Hasta 30 metros" },
      { key: "Expansión", value: "Auto-transmisión en cascada (mesh)" },
      { key: "Sincronización", value: "Automática entre múltiples controladores" },
      { key: "Modo No Molestar", value: "Evita encendido automático post-corte eléctrico" },
      { key: "Anti-Flicker", value: "Ajuste PWM para evitar parpadeo en cámaras" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "TE6 RGBW 102W",
    subtitle: "102W (8,5A) | Control por app y voz | WiFi + Bluetooth + RF | 1 año garantía",
    description: "Controlador inteligente con fuente de 102W / 8,5A. El modelo más vendido. Estándar para piletas residenciales. Conectividad triple WiFi + Bluetooth + RF 2.4GHz con integración Tuya Smart.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Modelo Más Vendido", description: "Equilibrio perfecto entre potencia y precio. El estándar para piletas residenciales." },
      { title: "Controlalo desde Tu Celular", description: "App Tuya Smart en iOS y Android. Cambiá colores, intensidad y efectos desde cualquier lugar." },
      { title: "Control por Voz", description: "Compatible con Amazon Alexa y Google Home." },
      { title: "Triple Conectividad", description: "WiFi + Bluetooth + RF. Siempre funciona aunque se caiga internet." },
      { title: "Expandí Sin Límites", description: "Auto-transmisión en cascada. Cobertura ilimitada." }
    ],

    main_specs: [
      { key: "Potencia", value: "102W", meaning: "Para piletas medianas/grandes" },
      { key: "Corriente", value: "8,5A", meaning: "Soporta más luces que TE3" },
      { key: "Conectividad", value: "WiFi + Bluetooth + RF", meaning: "Triple forma de control" },
      { key: "App", value: "Tuya Smart", meaning: "Control desde celular" },
      { key: "Voz", value: "Alexa + Google Home", meaning: "Comandos de voz" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "90~264VAC (rango universal)" },
      { key: "Tolerancia picos", value: "Hasta 300VAC por 5 segundos" },
      { key: "Salida", value: "12VDC · 8,5A · 102W" },
      { key: "Clase", value: "II (sin toma a tierra requerida)" },
      { key: "Carcasa", value: "Plástica encapsulada" },
      { key: "Protección IP", value: "IP67" },
      { key: "Protecciones activas", value: "Cortocircuito / Sobrecarga / Sobrevoltaje" },
      { key: "Conectividad", value: "WiFi 2.4GHz + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema", value: "Tuya Smart" },
      { key: "Compatibilidad voz", value: "Amazon Alexa / Google Home" },
      { key: "Alcance RF", value: "Hasta 30 metros" },
      { key: "Expansión", value: "Auto-transmisión en cascada (mesh)" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "TE9 RGBW 150W",
    subtitle: "150W (12,5A) | Protección 6KV | Carcasa metálica | 1 año garantía",
    description: "Controlador industrial con fuente de 150W / 12,5A para proyectos de gran formato. Chasis metálico Clase I con protección Surge 6KV/4KV y rango de entrada ultra amplio 100~305VAC.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Protección Industrial", description: "Surge Protection 6KV/4KV contra descargas atmosféricas y picos eléctricos. Blindaje total." },
      { title: "Carcasa Metálica Robusta", description: "Construcción de grado industrial. Disipación térmica superior. Mayor durabilidad." },
      { title: "Rango Ultra Amplio", description: "Funciona desde 100V hasta 305V sin problemas. Inmune a fluctuaciones de red." },
      { title: "Máxima Potencia", description: "150W / 12,5A. Para piletas grandes, hoteles, clubes. Alimentá más luces desde un solo punto." },
      { title: "Eficiencia 91,5%", description: "Menor desperdicio de energía. Menos calor generado. Mayor vida útil." }
    ],

    main_specs: [
      { key: "Potencia", value: "150W", meaning: "Para proyectos grandes" },
      { key: "Corriente", value: "12,5A", meaning: "Alta capacidad" },
      { key: "Rango voltaje", value: "100-305VAC", meaning: "Inmune a fluctuaciones" },
      { key: "Surge Protection", value: "6KV/4KV", meaning: "Protección contra rayos" },
      { key: "Eficiencia", value: "91,5%", meaning: "Ahorro energético" },
      { key: "Carcasa", value: "Metálica", meaning: "Grado industrial" },
      { key: "Clase", value: "I (con tierra)", meaning: "Mayor seguridad" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "100~305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC · 12,5A · 150W" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP", value: "IP67" },
      { key: "Clase", value: "I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Conectividad", value: "WiFi 2.4GHz + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema", value: "Tuya Smart" },
      { key: "Compatibilidad voz", value: "Amazon Alexa / Google Home" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "TE15 RGBW 300W Dual",
    subtitle: "300W DUAL (2x150W) | 2 salidas independientes | Protección 6KV | 1 año garantía",
    description: "Controlador industrial de doble salida independiente (2x150W = 300W totales). Permite alimentar dos circuitos desde un único gabinete. Ideal para proyectos de zonificación como piscina + spa.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Doble Salida Independiente", description: "2 circuitos de 150W cada uno desde UN SOLO equipo. Ideal para pileta + spa, o dos zonas separadas." },
      { title: "Ahorro en Instalación", description: "Un solo punto de conexión alimenta dos circuitos completos. Tablero más compacto." },
      { title: "Control Independiente", description: "Cada salida se controla por separado. Pileta en blanco y spa en colores simultáneamente." },
      { title: "Protección Industrial", description: "Surge 6KV/4KV. Rango 100-305VAC. Carcasa metálica." }
    ],

    main_specs: [
      { key: "Potencia total", value: "300W", meaning: "2 circuitos de 150W c/u" },
      { key: "Salidas", value: "2 independientes", meaning: "Pileta + spa / 2 zonas" },
      { key: "Surge Protection", value: "6KV/4KV", meaning: "Protección industrial" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "100~305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC × 2 · 12,5A × 2 · 300W totales" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP", value: "IP67" },
      { key: "Clase", value: "I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Conectividad", value: "WiFi 2.4GHz + Bluetooth + RF 2.4GHz" },
      { key: "Ecosistema", value: "Tuya Smart" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "TE3 W&W 60W",
    subtitle: "60W | Control RF | Sin WiFi | 1 año garantía",
    description: "Controlador CCT con fuente de 60W para proyectos pequeños de iluminación blanca ajustable. Control RF sin WiFi ni Bluetooth, con dimming de 20 niveles y memoria de estado.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Sin WiFi. Solo RF.", description: "Control 100% autónomo. No depende de internet ni routers. Ideal para instalaciones que priorizan simplicidad y confiabilidad." },
      { title: "11 Niveles de Temperatura de Color", description: "Desde blanco cálido (relajante) hasta blanco frío (brillante)." },
      { title: "20 Niveles de Intensidad", description: "Dimmea del 100% al 1% con precisión." },
      { title: "Memoria Automática", description: "Si se corta la luz, vuelve a la última configuración sin tocar nada." }
    ],

    main_specs: [
      { key: "Potencia", value: "60W", meaning: "Para piletas pequeñas/medianas" },
      { key: "Control", value: "RF (sin WiFi)", meaning: "Autónomo, no depende de internet" },
      { key: "CCT", value: "11 niveles", meaning: "Cálido a frío" },
      { key: "Dimming", value: "20 niveles", meaning: "Del 100% al 1%" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "90~264VAC" },
      { key: "Salida", value: "12VDC · 5A · 60W" },
      { key: "Clase", value: "II (sin toma a tierra)" },
      { key: "Protección IP", value: "IP67" },
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
    subtitle: "102W (8,5A) | Control RF | Sin WiFi | 1 año garantía",
    description: "Controlador CCT con fuente de 102W / 8,5A. Estándar residencial para piscinas con iluminación blanca ajustable. Control RF con dimming de 20 niveles y memoria de estado.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Sin WiFi. Solo RF.", description: "Control autónomo, sin depender de internet." },
      { title: "11 Niveles de Temperatura de Color", description: "De blanco cálido a blanco frío." },
      { title: "20 Niveles de Intensidad", description: "Precisión total del 100% al 1%." },
      { title: "Memoria Automática", description: "Recupera configuración tras corte de luz." }
    ],

    main_specs: [
      { key: "Potencia", value: "102W", meaning: "Para piletas medianas/grandes" },
      { key: "Corriente", value: "8,5A", meaning: "Más capacidad que TE3" },
      { key: "Control", value: "RF (sin WiFi)", meaning: "Autónomo" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "90~264VAC" },
      { key: "Salida", value: "12VDC · 8,5A · 102W" },
      { key: "Clase", value: "II (sin toma a tierra)" },
      { key: "Protección IP", value: "IP67" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles" },
      { key: "Memoria", value: "Resume function" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "TE9 W&W 150W",
    subtitle: "150W (12,5A) | Control RF | Carcasa metálica | 1 año garantía",
    description: "Controlador CCT industrial con fuente de 150W / 12,5A. Chasis metálico Clase I con Surge 6KV/4KV y PFC activo. Para hoteles, clubes y espejos de agua de gran envergadura.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Sin WiFi. Solo RF.", description: "Control autónomo de alta confiabilidad." },
      { title: "Protección Industrial", description: "Surge 6KV/4KV. Carcasa metálica. Para instalaciones exigentes." },
      { title: "Rango Ultra Amplio", description: "100-305VAC. Inmune a fluctuaciones." },
      { title: "Eficiencia 91,5%", description: "Máximo aprovechamiento energético." }
    ],

    main_specs: [
      { key: "Potencia", value: "150W", meaning: "Para proyectos grandes" },
      { key: "Corriente", value: "12,5A", meaning: "Alta capacidad" },
      { key: "Control", value: "RF (sin WiFi)", meaning: "Autónomo" },
      { key: "Surge Protection", value: "6KV/4KV", meaning: "Protección industrial" },
      { key: "Eficiencia", value: "91,5%", meaning: "Ahorro energético" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "100~305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC · 12,5A · 150W" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP", value: "IP67" },
      { key: "Clase", value: "I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles" },
      { key: "Memoria", value: "Resume function" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "TE15 W&W 300W Dual",
    subtitle: "300W DUAL (2x150W) | Control RF | Carcasa metálica | 1 año garantía",
    description: "Controlador CCT de doble salida independiente (2x150W = 300W totales). Ideal para proyectos de zonificación con luz blanca ajustable. Chasis metálico industrial con Surge 6KV/4KV.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Doble Salida Independiente", description: "2 circuitos de 150W. Pileta + spa o dos zonas separadas." },
      { title: "Sin WiFi. Solo RF.", description: "Control autónomo sin internet." },
      { title: "Protección Industrial", description: "Surge 6KV/4KV. Carcasa metálica." }
    ],

    main_specs: [
      { key: "Potencia total", value: "300W", meaning: "2 circuitos de 150W c/u" },
      { key: "Salidas", value: "2 independientes", meaning: "2 zonas" },
      { key: "Control", value: "RF (sin WiFi)", meaning: "Autónomo" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "100~305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "91,5%" },
      { key: "Salida", value: "12VDC × 2 · 12,5A × 2 · 300W totales" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Clase", value: "I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Transmisión", value: "Radiofrecuencia RF (sin WiFi ni Bluetooth)" },
      { key: "CCT", value: "Blanco cálido a blanco frío en 11 niveles" },
      { key: "Dimming", value: "100% al 1% en 20 niveles" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  },

  {
    name: "PCE-100-SYS 24V 96W",
    subtitle: "96W (24VDC) | Compatible OSIRE | RF 2.4GHz | 1 año garantía",
    description: "Controlador SYS con fuente 24VDC para uso exclusivo con luminarias OSIRE. Control RF 2.4GHz con auto-transmisión en cascada para cobertura ilimitada. Requiere Gateway separado para control por app.",
    base_price: 0.0,
    category: "Controladores",
    brand: "Pooled",
    images: [],
    variants: [],

    benefits: [
      { title: "Exclusivo para OSIRE 24V", description: "Diseñado específicamente para la línea premium OSIRE. Optimización total." },
      { title: "Protección Industrial Máxima", description: "Surge Protection 6KV/4KV. Rango 100-305VAC. Carcasa metálica." },
      { title: "Control RF de Largo Alcance", description: "Señal RF hasta 30 metros. Auto-transmisión en cascada para cobertura ilimitada." },
      { title: "Sincronización Automática", description: "Múltiples controladores se sincronizan perfectamente." },
      { title: "App Opcional", description: "Agregá Gateway WiFi (vendido separado) para controlar desde el celular." }
    ],

    main_specs: [
      { key: "Potencia", value: "96W", meaning: "Para sistema OSIRE" },
      { key: "Voltaje salida", value: "24VDC", meaning: "SOLO compatible con OSIRE" },
      { key: "Control", value: "RF 2.4GHz", meaning: "Hasta 30m de alcance" },
      { key: "WiFi", value: "Opcional (Gateway)", meaning: "Se vende por separado" },
      { key: "Surge Protection", value: "6KV/4KV", meaning: "Protección industrial" },
      { key: "Eficiencia", value: "92%", meaning: "Máxima eficiencia" },
      { key: "Garantía", value: "1 año", meaning: "Respaldo comercial" }
    ],

    specs: [
      { key: "Entrada", value: "100~305VAC · 47~63Hz" },
      { key: "Eficiencia", value: "92%" },
      { key: "Salida", value: "24VDC · 4A · 96W" },
      { key: "Protecciones", value: "SCP / OCP / OVP / OTP" },
      { key: "Surge Protection", value: "6KV (Línea-Tierra) · 4KV (Línea-Línea)" },
      { key: "Protección IP", value: "IP67" },
      { key: "Clase", value: "I (con toma a tierra requerida)" },
      { key: "Carcasa", value: "Metálica encapsulada" },
      { key: "Control", value: "RF inalámbrico 2.4GHz" },
      { key: "Alcance", value: "Hasta 30 metros" },
      { key: "Control por App", value: "Sí (requiere Gateway 2.4GHz separado, vendido aparte)" },
      { key: "Auto-transmisión cascada", value: "Sí — cobertura ilimitada" },
      { key: "Sincronización", value: "Automática entre múltiples controladores" },
      { key: "Compatibilidad", value: "Exclusivo para luminarias 24V (OSIRE)" },
      { key: "Garantía", value: "1 año" }
    ],

    created_at: now,
    updated_at: now
  }

];

db.products.insertMany(products);
print(`✅ ${products.length} productos insertados`);

// ============================================================
// ÍNDICES (para búsquedas rápidas)
// ============================================================
db.products.createIndex({ category: 1 });
db.products.createIndex({ brand: 1 });
db.products.createIndex({ name: "text", description: "text" });
print("✅ Índices creados");

// ============================================================
// RESUMEN
// ============================================================
print("\n📊 RESUMEN:");
print(`   users:    ${db.users.countDocuments()}`);
print(`   products: ${db.products.countDocuments()}`);
print("\n⚠️  PENDIENTE:");
print("   - Precios (base_price en 0.0)");
print("   - URLs de imágenes (images: [])");
print("\n✅ Seed v2 completado exitosamente");
