![Next Gen](https://i.imgur.com/xL3cCNv.png)
A project generator for my next template!
The whole point of the next template was to make it easy and fast to start making apps, however it was always a pain creating multiple git repos for all my projects and modifying all the files to set up the project correctly making sure you don't miss any config files.

I've finaly got around to updating this to nextjs 13, Expect your build times to improve massivly. 

# More about my Template

## [Read about it here](https://github.com/AndreCox/next-template)

### Key Features

1. 🗳️ MobX State Management
2. 📱 Fully cross platform, you can create your web app, then deploy to both IOS and Android
3. 🪶 Comes with Tailwindcss by default; no more thinking up css class names while still being lightweight
4. 📄 Github Pages support, simply push your code and your website will be automatically deployed.
5. ⏭️ Next Js seriously makes development way easier. The major update from previous template.
6. 🖥️ NEW! Tauri Support Build for Windows Mac and Linux

# How to use

## New Method

Now the program is published you can install the software with

```
yarn global add nxgen
```

Or

```
npm install -g nxgen
```

Once it has been installed you may need to restart your terminal, then you can type `nextgen`

## Old Method

Simply download the appropriate version from the releases tab and add it to your path. Then in an empty foler simply type.

```
nextgen
```

##

Then the program will guide you through installation.

You will be prompted for some information about your product as follows.

1. Your project name.
2. A description of your project.
3. Your Name.
4. A project id (com.company.app).

After this it will modify all the necesarry files.
Next if the software detects you have git installed verson control will be automatically set up.
Finally it will install all the necessary dependencies for you if Yarn is detected.

## A note on Windows Defender

Right now Go Programs are often detected as malware, This however is a false positive. You can read the source yourself to confirm that there is nothing malicious inside. If you're really paranoid you can just use the template directly and modify all of the parameters manualy.
