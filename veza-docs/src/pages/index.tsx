import React, {type ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <div className="row">
          <div className="col col--8">
        <Heading as="h1" className="hero__title">
              🎵 {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
            <p className="hero__description">
              Une plateforme moderne de streaming audio avec chat en temps réel, 
              construite avec Go, Rust et React. Performance, scalabilité et expérience utilisateur au rendez-vous.
            </p>
        <div className={styles.buttons}>
              <Link
                className="button button--primary button--lg"
                to="/docs/">
                📚 Commencer la Documentation
              </Link>
          <Link
            className="button button--secondary button--lg"
                to="https://github.com/okinrev/veza-full-stack">
                🚀 Voir sur GitHub
          </Link>
            </div>
          </div>
          <div className="col col--4">
            <div className={styles.heroImage}>
              <div className={styles.heroCard}>
                <div className={styles.heroCardHeader}>
                  <h3>🎯 Services</h3>
                </div>
                <div className={styles.heroCardBody}>
                  <ul>
                    <li>🎵 Streaming Audio</li>
                    <li>💬 Chat Temps Réel</li>
                    <li>🔐 Authentification</li>
                    <li>📊 Analytics</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          <div className="col col--4">
            <div className="card">
              <div className="card__header">
                <h3>🏗️ Architecture Moderne</h3>
              </div>
              <div className="card__body">
                <p>
                  Architecture microservices avec Go, Rust et React. 
                  Performance optimale et scalabilité garantie.
                </p>
              </div>
              <div className="card__footer">
                <Link to="/docs/architecture/backend-architecture" className="button button--primary button--block">
                  Voir l'Architecture
                </Link>
              </div>
            </div>
          </div>
          
          <div className="col col--4">
            <div className="card">
              <div className="card__header">
                <h3>🎵 Streaming Audio</h3>
              </div>
              <div className="card__body">
                <p>
                  Support de multiples formats audio avec compression adaptative. 
                  Qualité optimale selon la bande passante.
                </p>
              </div>
              <div className="card__footer">
                <Link to="/docs/stream-server/src/main" className="button button--primary button--block">
                  Voir le Stream Server
                </Link>
              </div>
            </div>
          </div>
          
          <div className="col col--4">
            <div className="card">
              <div className="card__header">
                <h3>💬 Chat Temps Réel</h3>
              </div>
              <div className="card__body">
                <p>
                  Communication instantanée avec WebSocket. 
                  Salons privés, messages directs et modération automatique.
                </p>
              </div>
              <div className="card__footer">
                <Link to="/docs/chat-server/src/main" className="button button--primary button--block">
                  Voir le Chat Server
                </Link>
              </div>
            </div>
          </div>
        </div>
        
        <div className="row" style={{marginTop: '2rem'}}>
          <div className="col col--6">
            <div className="card">
              <div className="card__header">
                <h3>🔌 API REST Complète</h3>
              </div>
              <div className="card__body">
                <p>
                  API REST moderne avec authentification JWT, rate limiting 
                  et documentation OpenAPI complète.
                </p>
                <ul>
                  <li>Authentification JWT & OAuth2</li>
                  <li>Gestion des utilisateurs</li>
                  <li>Upload de fichiers</li>
                  <li>Webhooks</li>
                </ul>
              </div>
              <div className="card__footer">
                <Link to="/docs/api/endpoints-reference" className="button button--primary button--block">
                  Voir l'API
                </Link>
              </div>
            </div>
          </div>
          
          <div className="col col--6">
            <div className="card">
              <div className="card__header">
                <h3>🚀 Déploiement Simple</h3>
              </div>
              <div className="card__body">
                <p>
                  Déploiement automatisé avec Docker, Kubernetes et CI/CD. 
                  Monitoring et observabilité intégrés.
                </p>
                <ul>
                  <li>Docker Compose</li>
                  <li>Kubernetes</li>
                  <li>CI/CD GitHub Actions</li>
                  <li>Monitoring Prometheus</li>
                </ul>
              </div>
              <div className="card__footer">
                <Link to="/docs/deployment/deployment-guide" className="button button--primary button--block">
                  Voir le Guide
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function HomepageStats() {
  return (
    <section className={styles.stats}>
      <div className="container">
        <div className="row">
          <div className="col col--3">
            <div className={styles.statItem}>
              <h3>3</h3>
              <p>Services</p>
            </div>
          </div>
          <div className="col col--3">
            <div className={styles.statItem}>
              <h3>100%</h3>
              <p>Documenté</p>
            </div>
          </div>
          <div className="col col--3">
            <div className={styles.statItem}>
              <h3>24/7</h3>
              <p>Monitoring</p>
            </div>
          </div>
          <div className="col col--3">
            <div className={styles.statItem}>
              <h3>∞</h3>
              <p>Scalable</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title} - ${siteConfig.tagline}`}
      description="Documentation complète de la plateforme Veza - Streaming audio et chat en temps réel avec Go, Rust et React">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <HomepageStats />
      </main>
    </Layout>
  );
}
