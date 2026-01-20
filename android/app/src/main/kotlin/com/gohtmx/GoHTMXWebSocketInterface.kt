package com.gohtmx

import android.webkit.JavascriptInterface
import gohtmx.Gohtmx

/**
 * JavaScript interface for virtual WebSocket connections.
 * Exposed as GoHTMXNative in JavaScript.
 */
class GoHTMXWebSocketInterface(private val activity: GoHTMXActivity) : Gohtmx.WebSocketCallback {

    private val activeSessions = mutableSetOf<String>()

    init {
        Gohtmx.setWebSocketCallback(this)
    }

    /**
     * Connect to a virtual WebSocket.
     * Called from JavaScript: GoHTMXNative.wsConnect(url)
     */
    @JavascriptInterface
    fun wsConnect(url: String): String {
        return try {
            val sessionID = Gohtmx.webSocketConnect(url)
            activeSessions.add(sessionID)
            sessionID
        } catch (e: Exception) {
            ""
        }
    }

    /**
     * Send a message through a virtual WebSocket.
     * Called from JavaScript: GoHTMXNative.wsSend(sessionID, data)
     */
    @JavascriptInterface
    fun wsSend(sessionID: String, data: String): String? {
        return try {
            Gohtmx.webSocketSend(sessionID, data)
        } catch (e: Exception) {
            null
        }
    }

    /**
     * Close a virtual WebSocket connection.
     * Called from JavaScript: GoHTMXNative.wsClose(sessionID)
     */
    @JavascriptInterface
    fun wsClose(sessionID: String) {
        try {
            Gohtmx.webSocketClose(sessionID)
        } catch (e: Exception) {
            // Ignore errors on close
        }
        activeSessions.remove(sessionID)
    }

    // WebSocketCallback implementation

    override fun onMessage(sessionID: String?, data: String?) {
        if (sessionID == null || data == null) return

        val escaped = data
            .replace("\\", "\\\\")
            .replace("'", "\\'")
            .replace("\n", "\\n")
            .replace("\r", "\\r")

        activity.runOnUiThread {
            activity.evaluateJavaScript(
                "window._gohtmx_ws_message('$sessionID', '$escaped')"
            )
        }
    }

    override fun onClose(sessionID: String?, code: Long, reason: String?) {
        if (sessionID == null) return

        activeSessions.remove(sessionID)

        val reasonEscaped = (reason ?: "").replace("'", "\\'")

        activity.runOnUiThread {
            activity.evaluateJavaScript(
                "window._gohtmx_ws_close('$sessionID', $code, '$reasonEscaped')"
            )
        }
    }

    override fun onError(sessionID: String?, error: String?) {
        if (sessionID == null || error == null) return

        val errorEscaped = error.replace("'", "\\'")

        activity.runOnUiThread {
            activity.evaluateJavaScript(
                "window._gohtmx_ws_error('$sessionID', '$errorEscaped')"
            )
        }
    }

    /**
     * Close all active sessions
     */
    fun closeAll() {
        activeSessions.toList().forEach { wsClose(it) }
    }
}
