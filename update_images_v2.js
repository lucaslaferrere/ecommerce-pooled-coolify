const db = connect("mongodb://root:Q62XCH20VplJvaCzy6VM0I53mIzmrZs4wbwNpfJBwCcOyEix7Up7AzEt7zmw9pSH@m48koc8ckkkkg088c84o8880:27017/ecommerce?directConnection=true&authSource=admin");

const BASE = "https://pub-6232b7116b2042bbb2308cbdd5eec1b1.r2.dev/Pooled%20-%20Imagenes";

const updates = [
  {
    name: "OSIRE 18W RGBW",
    images: [
      `${BASE}/Osire%201.png`,
      `${BASE}/Osire%202.png`,
      `${BASE}/Osire%203.png`,
      `${BASE}/Osire%204.png`,
      `${BASE}/Osire%205.png`,
      `${BASE}/Osire6.png`,
      `${BASE}/Osire%207.png`,
      `${BASE}/Osire%208.png`,
    ]
  },
  {
    name: "HORUS 18W RGBW",
    images: [
      `${BASE}/Horus%201.png`,
      `${BASE}/Horus%202.png`,
      `${BASE}/Horus%203.png`,
    ]
  },
  {
    name: "NAZAR 9W RGBW",
    images: [
      `${BASE}/Nazar%201.jpg`,
      `${BASE}/Nazar%202.jpg`,
      `${BASE}/Nazar%203.jpg`,
      `${BASE}/Nazar%204.jpg`,
    ]
  },
];

let updated = 0;
for (const u of updates) {
  const res = db.products.updateOne(
    { name: u.name },
    { $set: { images: u.images, updated_at: new Date() } }
  );
  if (res.matchedCount > 0) {
    print(`✅ ${u.name} → ${u.images.length} imágenes`);
    updated++;
  } else {
    print(`⚠️  No encontrado: ${u.name}`);
  }
}

print(`\n✅ ${updated}/${updates.length} productos actualizados`);
