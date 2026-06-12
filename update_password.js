const db = connect("mongodb://root:Q62XCH20VplJvaCzy6VM0I53mIzmrZs4wbwNpfJBwCcOyEix7Up7AzEt7zmw9pSH@m48koc8ckkkkg088c84o8880:27017/ecommerce?directConnection=true&authSource=admin");

const result = db.users.updateOne(
  { email: "admin@pooled.com.ar" },
  { $set: { password_hash: "$2a$10$QwpKSObIMAXHFQsPR2PToui2t6LPsH9aHvc25Dxa8hFrHAiPrSsbq" } }
);

print(result.modifiedCount === 1 ? "✅ Password actualizada" : "⚠️  No se encontró el usuario");
