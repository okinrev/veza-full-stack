import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

/**
 * Configuration de la sidebar pour la documentation Veza
 * 
 * Cette sidebar organise la documentation en sections logiques :
 * - Vue d'ensemble et architecture
 * - Services (Backend, Chat, Stream)
 * - API et développement
 * - Déploiement et opérations
 * - Guides et références
 */
const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    // Page d'accueil de la documentation
    'index',
    
    // Vue d'ensemble et architecture
    {
      type: 'category',
      label: '🏗️ Architecture',
      link: {
        type: 'doc',
        id: 'architecture/backend-architecture',
      },
      items: [
        'architecture/backend-architecture',
        'architecture/chat-server-architecture',
        'architecture/stream-server-architecture',
        'architecture/backend-config',
      ],
    },
    
    // Services principaux
    {
      type: 'category',
      label: '🚀 Services',
      items: [
        {
          type: 'category',
          label: 'Backend API',
          items: [
            'backend-api/src/cmd-server-main',
            'backend-api/src/config-config',
          ],
        },
        {
          type: 'category',
          label: 'Chat Server',
          items: [
            'chat-server/src/main',
          ],
        },
        {
          type: 'category',
          label: 'Stream Server',
          items: [
            'stream-server/src/main',
          ],
        },
      ],
    },
    
    // API et développement
    {
      type: 'category',
      label: '🔌 API & Développement',
      items: [
        'api/endpoints-reference',
        'api/backend-api',
        'database/schema',
      ],
    },
    
    // Déploiement et opérations
    {
      type: 'category',
      label: '🚀 Déploiement',
      items: [
        'deployment/deployment-guide',
        'deployment/guide',
      ],
    },
    
    // Diagrammes et visualisations
    {
      type: 'category',
      label: '📊 Diagrammes',
      items: [
        'diagrams/architecture-overview',
        'diagrams/data-flow',
      ],
    },
    
    // Modules et dépendances
    {
      type: 'category',
      label: '🔗 Modules',
      items: [
        'modules/dependency-map',
      ],
    },
  ],
};

export default sidebars;
