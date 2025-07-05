"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var prism_react_renderer_1 = require("prism-react-renderer");
// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)
var config = {
    title: 'Veza Documentation',
    tagline: 'Plateforme de streaming audio et chat en temps réel',
    favicon: 'img/favicon.ico',
    themes: ['@docusaurus/theme-mermaid'],
    // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
    future: {
        v4: true, // Improve compatibility with the upcoming Docusaurus v4
    },
    // Set the production url of your site here
    url: 'https://veza-docs.vercel.app',
    // Set the /<baseUrl>/ pathname under which your site is served
    // For GitHub pages deployment, it is often '/<projectName>/'
    baseUrl: '/',
    // GitHub pages deployment config.
    // If you aren't using GitHub pages, you don't need these.
    organizationName: 'okinrev', // Usually your GitHub org/user name.
    projectName: 'veza-full-stack', // Usually your repo name.
    onBrokenLinks: 'warn',
    onBrokenMarkdownLinks: 'warn',
    // Even if you don't use internationalization, you can use this field to set
    // useful metadata like html lang. For example, if your site is Chinese, you
    // may want to replace "en" with "zh-Hans".
    i18n: {
        defaultLocale: 'fr',
        locales: ['fr'],
    },
    presets: [
        [
            '@docusaurus/preset-classic',
            {
                docs: {
                    sidebarPath: require.resolve('./sidebars.ts'),
                    editUrl: 'https://github.com/okinrev/veza-full-stack/tree/main/veza-docs/',
                },
                blog: {
                    showReadingTime: true,
                    feedOptions: {
                        type: ['rss', 'atom'],
                        xslt: true,
                    },
                    editUrl: 'https://github.com/okinrev/veza-full-stack/tree/main/veza-docs/',
                    onInlineTags: 'warn',
                    onInlineAuthors: 'warn',
                    onUntruncatedBlogPosts: 'warn',
                },
                theme: {
                    customCss: require.resolve('./src/css/custom.css'),
                },
            },
        ],
    ],
    themeConfig: {
        // Replace with your project's social card
        image: 'img/veza-social-card.jpg',
        navbar: {
            title: 'Veza Docs',
            logo: {
                alt: 'Veza Logo',
                src: 'img/veza-logo.svg',
            },
            items: [
                {
                    type: 'docSidebar',
                    sidebarId: 'tutorialSidebar',
                    position: 'left',
                    label: 'Documentation',
                },
                { to: '/blog', label: 'Blog', position: 'left' },
                {
                    href: 'https://github.com/okinrev/veza-full-stack',
                    label: 'GitHub',
                    position: 'right',
                },
            ],
        },
        footer: {
            style: 'dark',
            links: [
                {
                    title: 'Documentation',
                    items: [
                        {
                            label: 'Architecture',
                            to: '/docs/architecture/backend-architecture',
                        },
                        {
                            label: 'API Reference',
                            to: '/docs/api/endpoints-reference',
                        },
                        {
                            label: 'Déploiement',
                            to: '/docs/deployment/deployment-guide',
                        },
                    ],
                },
                {
                    title: 'Services',
                    items: [
                        {
                            label: 'Backend API',
                            to: '/docs/backend-api/src/cmd-server-main',
                        },
                        {
                            label: 'Chat Server',
                            to: '/docs/chat-server/src/main',
                        },
                        {
                            label: 'Stream Server',
                            to: '/docs/stream-server/src/main',
                        },
                    ],
                },
                {
                    title: 'Ressources',
                    items: [
                        {
                            label: 'GitHub',
                            href: 'https://github.com/okinrev/veza-full-stack',
                        },
                        {
                            label: 'Issues',
                            href: 'https://github.com/okinrev/veza-full-stack/issues',
                        },
                        {
                            label: 'Discussions',
                            href: 'https://github.com/okinrev/veza-full-stack/discussions',
                        },
                    ],
                },
            ],
            copyright: "Copyright \u00A9 ".concat(new Date().getFullYear(), " Veza. Built with Docusaurus."),
        },
        prism: {
            theme: prism_react_renderer_1.themes.github,
            darkTheme: prism_react_renderer_1.themes.dracula,
            additionalLanguages: ['bash', 'json', 'yaml', 'toml', 'rust', 'go'],
        },
        mermaid: {
            theme: { light: 'default', dark: 'dark' },
            options: {
                maxTextSize: 50000,
                securityLevel: 'loose',
            },
        },
        // Configuration pour un design moderne
        colorMode: {
            defaultMode: 'light',
            disableSwitch: false,
            respectPrefersColorScheme: true,
        },
        // Métadonnées pour SEO
        metadata: [
            { name: 'keywords', content: 'veza, streaming, audio, chat, real-time, documentation' },
            { name: 'description', content: 'Documentation complète de la plateforme Veza - Streaming audio et chat en temps réel' },
        ],
    },
    markdown: {
        mermaid: true,
    },
};
exports.default = config;
