package com.gohtmx

import android.webkit.WebView
import gohtmx.Gohtmx
import org.json.JSONObject

/**
 * Main bridge class that connects Android to the Go framework
 */
object GoHTMXBridge {

    private var webView: WebView? = null

    /**
     * Initialize the Go bridge. Call once at app startup.
     */
    fun initialize() {
        Gohtmx.initialize()
    }

    /**
     * Configure the bridge with a WebView
     */
    fun configure(webView: WebView) {
        this.webView = webView
    }

    /**
     * Check if the bridge is ready
     */
    val isReady: Boolean
        get() = Gohtmx.isReady()

    /**
     * Handle an HTTP request and return the response
     */
    fun handleRequest(
        method: String,
        url: String,
        headers: Map<String, String> = emptyMap(),
        body: ByteArray? = null
    ): GoHTMXResponse {
        val headersJson = JSONObject(headers).toString()
        val response = Gohtmx.handleRequest(method, url, headersJson, body)
            ?: return GoHTMXResponse(500, emptyMap(), ByteArray(0))

        return GoHTMXResponse.from(response)
    }

    /**
     * Get the initial HTML page content
     */
    fun renderInitialPage(): String {
        return Gohtmx.renderInitialPage()
    }

    /**
     * Shutdown the bridge
     */
    fun shutdown() {
        Gohtmx.shutdown()
    }
}

/**
 * Kotlin-friendly response wrapper
 */
data class GoHTMXResponse(
    val status: Int,
    val headers: Map<String, String>,
    val body: ByteArray
) {
    val bodyString: String
        get() = String(body, Charsets.UTF_8)

    companion object {
        fun from(response: gohtmx.Core.Response): GoHTMXResponse {
            val headers = mutableMapOf<String, String>()
            try {
                val headersJson = JSONObject(response.headers)
                headersJson.keys().forEach { key ->
                    headers[key] = headersJson.getString(key)
                }
            } catch (e: Exception) {
                // Ignore JSON parsing errors
            }

            return GoHTMXResponse(
                status = response.status.toInt(),
                headers = headers,
                body = response.body ?: ByteArray(0)
            )
        }
    }

    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false
        other as GoHTMXResponse
        return status == other.status && headers == other.headers && body.contentEquals(other.body)
    }

    override fun hashCode(): Int {
        var result = status
        result = 31 * result + headers.hashCode()
        result = 31 * result + body.contentHashCode()
        return result
    }
}
