const express = require('express');
const app = express();

// Google OAuth2 Callback Endpoint
app.get('/oauth2/callback', (req, res) => {
  const authorizationCode = req.query.code;
  console.log('Authorization Code:', authorizationCode);

  // 您可以在此處處理交換 Access Token 的邏輯
  res.send('OAuth authentication successful! You can close this tab.');
});

app.listen(8080, () => {
  console.log('Server is running on http://localhost:8080');
});
