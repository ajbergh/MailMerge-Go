/**
 * Application Entry Point
 * 
 * This is the main entry point for the React frontend application.
 * It renders the root App component into the DOM using React 18's
 * createRoot API with StrictMode enabled for development checks.
 * 
 * The application runs inside a Wails WebView, communicating with
 * the Go backend through generated bindings in the wailsjs directory.
 */
import React from 'react'
import {createRoot} from 'react-dom/client'
import App from './App'

// Get the root DOM element
const container = document.getElementById('root')

// Create React 18 root
const root = createRoot(container!)

// Render the application with StrictMode for development checks
root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)
