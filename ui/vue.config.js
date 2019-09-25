const FaviconsWebpackPlugin = require('favicons-webpack-plugin');

// vue.config.js
module.exports = {
  configureWebpack: {
    plugins: [
      new FaviconsWebpackPlugin({
        logo: './src/assets/images/logo.png',
        inject: true,
        title: 'Adagio'
      })
    ]
  }
}
