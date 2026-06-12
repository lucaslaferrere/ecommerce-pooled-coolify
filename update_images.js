// ============================================================
// UPDATE IMAGES - POOLED ECOMMERCE
// Correr con SSH + docker exec igual que el seed
// ============================================================

const db = connect("mongodb://root:Q62XCH20VplJvaCzy6VM0I53mIzmrZs4wbwNpfJBwCcOyEix7Up7AzEt7zmw9pSH@m48koc8ckkkkg088c84o8880:27017/ecommerce?directConnection=true&authSource=admin");

const base = "https://pub-6232b7116b2042bbb2308cbdd5eec1b1.r2.dev/Pooled%20-%20Imagenes";

const updates = [
  { name: "OSIRE RGBW 18W",            images: [`${base}/Osire.png`, `${base}/Osire2.png`, `${base}/Osire3.png`] },
  { name: "HORUS RGBW 18W",            images: [`${base}/HORUS%20RGBW.png`, `${base}/HORUS%20RGBW2.png`] },
  { name: "NAZAR RGBW 9W",             images: [`${base}/NAZAR%20RGBW.png`] },
  { name: "HORUS W&W 7.2W",            images: [`${base}/horus%20w%26w.png`] },
  { name: "TE3 RGBW 60W",              images: [`${base}/TE3%20RGBW%20%20CONTROLADOR.png`] },
  { name: "TE6 RGBW 102W",             images: [`${base}/TE6%20RGBW%20%20CONTROLADOR.png`] },
  { name: "TE9 RGBW 150W",             images: [`${base}/TE9%20RGBW%20%20CONTROLADOR.png`] },
  { name: "PCE-60-12 Controlador 5 en 1", images: [`${base}/PCE-60-12.png`] },
];

let updated = 0;
for (const u of updates) {
  const result = db.products.updateOne({ name: u.name }, { $set: { images: u.images } });
  if (result.modifiedCount > 0) {
    print(`✅ ${u.name}`);
    updated++;
  } else {
    print(`⚠️  No encontrado: ${u.name}`);
  }
}

print(`\n📊 ${updated}/${updates.length} productos actualizados`);
