// Copyright 2019 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

"use strict";

const path = require("path");
const rimraf = require("rimraf");
const webpack = require("webpack");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const VisualizerPlugin = require("webpack-visualizer-plugin");

// Remove a broken dependency that Yarn insists upon installing before every
// Webpack compile. We also do this when installing dependencies via Make, but
// it"s common to run e.g. `yarn add` manually without re-running Make, which
// will reinstall the broken dependency. The error this dependency causes is
// horribly cryptic, so it"s important to remove it aggressively.
//
// See: https://github.com/yarnpkg/yarn/issues/2987
class RemoveBrokenDependenciesPlugin {
  apply(compiler) {
    compiler.plugin("compile", () => rimraf.sync("./node_modules/@types/node"));
  }
}

let DashboardPlugin;
try {
  DashboardPlugin = require("./opt/node_modules/webpack-dashboard/plugin");
} catch (e) {
  DashboardPlugin = class { apply() { /* no-op */ } };
}

const proxyPrefixes = ["/_admin", "/_status", "/ts", "/login", "/logout"];
function shouldProxy(reqPath) {
  if (reqPath === "/") {
    return true;
  }
  return proxyPrefixes.some((prefix) => (
    reqPath.startsWith(prefix)
  ));
}

// tslint:disable:object-literal-sort-keys
module.exports = (env) => {
  let localRoots = [path.resolve(__dirname)];
  if (env.dist === "ccl") {
    // CCL modules shadow OSS modules.
    localRoots.unshift(path.resolve(__dirname, "ccl"));
  }

  return {
    entry: ["./src/index.tsx"],
    output: {
      filename: "bundle.js",
      path: path.resolve(__dirname, `dist${env.dist}`),
    },

    resolve: {
      // Add resolvable extensions.
      extensions: [".ts", ".tsx", ".js", ".json", ".styl", ".css"],
      // First check for local modules, then for third-party modules from
      // node_modules.
      //
      // These module roots are transformed into absolute paths, by
      // path.resolve, to ensure that only the exact directory is checked.
      // Relative paths would trigger the resolution behavior used by Node.js
      // for "node_modules", i.e., checking for a "node_modules" directory in
      // the current directory *or any parent directory*.
      modules: [
        ...localRoots,
        path.resolve(__dirname, "node_modules"),
      ],
      alias: {oss: path.resolve(__dirname)},
    },

    module: {
      rules: [
        { test: /\.css$/, use: [ "style-loader", "css-loader" ] },
        {
          test: /\.styl$/,
          use: [
            "cache-loader",
            "style-loader",
            "css-loader",
            {
              loader: "stylus-loader",
              options: {
                use: [require("nib")()],
              },
            },
          ],
        },
        {
          test: /\.(png|jpg|gif|svg|eot|ttf|woff|woff2)$/,
          loader: "url-loader",
          options: {
            limit: 10000,
          },
        },
        { test: /\.html$/, loader: "file-loader" },
        {
          test: /\.js$/,
          include: localRoots,
          use: ["cache-loader", "babel-loader"],
        },
        {
          test: /\.tsx?$/,
          include: localRoots,
          use: [
            "cache-loader",
            "babel-loader",
            { loader: "ts-loader", options: { happyPackMode: true } },
          ],
        },

        // All output ".js" files will have any sourcemaps re-processed by "source-map-loader".
        { enforce: "pre", test: /\.js$/, loader: "source-map-loader" },
      ],
    },

    plugins: [
      new RemoveBrokenDependenciesPlugin(),
      // See "DLLs for speedy builds" in the README for details.
      new webpack.DllReferencePlugin({
        manifest: require(`./protos.${env.dist}.manifest.json`),
      }),
      new webpack.DllReferencePlugin({
        manifest: require("./vendor.oss.manifest.json"),
      }),
      new CopyWebpackPlugin([{ from: "favicon.ico", to: "favicon.ico" }]),
      new DashboardPlugin(),
      new VisualizerPlugin({ filename: `../dist/stats.${env.dist}.html` }),
    ],

    // https://webpack.js.org/configuration/stats/
    stats: {
      colors: true,
      chunks: false,
    },

    devServer: {
      contentBase: path.join(__dirname, `dist${env.dist}`),
      index: "",
      proxy: {
        // Note: this shouldn't require a custom bypass function to work;
        // docs say that setting `index: ''` is sufficient to proxy `/`.
        // However, that did not work, and may require upgrading to webpack 4.x.
        "/": {
          secure: false,
          target: process.env.TARGET,
          bypass: (req) => {
            if (shouldProxy(req.path)) {
              return false;
            }
            return req.path;
          },
        },
      },
    },
  };
};
