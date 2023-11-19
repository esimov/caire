// SPDX-License-Identifier: Unlicense OR MIT

package org.gioui;

import java.lang.Class;
import java.lang.IllegalAccessException;
import java.lang.InstantiationException;
import java.lang.ExceptionInInitializerError;
import java.lang.SecurityException;
import android.app.Activity;
import android.app.Fragment;
import android.app.FragmentManager;
import android.app.FragmentTransaction;
import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Color;
import android.graphics.Matrix;
import android.graphics.Rect;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.SystemClock;
import android.text.TextUtils;
import android.text.Selection;
import android.text.SpannableStringBuilder;
import android.util.AttributeSet;
import android.util.TypedValue;
import android.view.Choreographer;
import android.view.Display;
import android.view.KeyCharacterMap;
import android.view.KeyEvent;
import android.view.MotionEvent;
import android.view.PointerIcon;
import android.view.View;
import android.view.ViewConfiguration;
import android.view.WindowInsets;
import android.view.Surface;
import android.view.SurfaceView;
import android.view.SurfaceHolder;
import android.view.Window;
import android.view.WindowInsetsController;
import android.view.WindowManager;
import android.view.inputmethod.CorrectionInfo;
import android.view.inputmethod.CompletionInfo;
import android.view.inputmethod.CursorAnchorInfo;
import android.view.inputmethod.EditorInfo;
import android.view.inputmethod.ExtractedText;
import android.view.inputmethod.ExtractedTextRequest;
import android.view.inputmethod.InputConnection;
import android.view.inputmethod.InputMethodManager;
import android.view.inputmethod.InputContentInfo;
import android.view.inputmethod.SurroundingText;
import android.view.accessibility.AccessibilityNodeProvider;
import android.view.accessibility.AccessibilityNodeInfo;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityManager;

import java.io.UnsupportedEncodingException;

public final class GioView extends SurfaceView implements Choreographer.FrameCallback {
	private static boolean jniLoaded;

	private final SurfaceHolder.Callback surfCallbacks;
	private final View.OnFocusChangeListener focusCallback;
	private final InputMethodManager imm;
	private final float scrollXScale;
	private final float scrollYScale;
	private int keyboardHint;
	private AccessibilityManager accessManager;

	private long nhandle;

	public GioView(Context context) {
		this(context, null);
	}

	public GioView(Context context, AttributeSet attrs) {
		super(context, attrs);
		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
			setSystemUiVisibility(View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION | View.SYSTEM_UI_FLAG_LAYOUT_STABLE);
		}
		setLayoutParams(new WindowManager.LayoutParams(WindowManager.LayoutParams.MATCH_PARENT, WindowManager.LayoutParams.MATCH_PARENT));

		// Late initialization of the Go runtime to wait for a valid context.
		Gio.init(context.getApplicationContext());

		// Set background color to transparent to avoid a flickering
		// issue on ChromeOS.
		setBackgroundColor(Color.argb(0, 0, 0, 0));

		ViewConfiguration conf = ViewConfiguration.get(context);
		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
			scrollXScale = conf.getScaledHorizontalScrollFactor();
			scrollYScale = conf.getScaledVerticalScrollFactor();

			// The platform focus highlight is not aware of Gio's widgets.
			setDefaultFocusHighlightEnabled(false);
		} else {
			float listItemHeight = 48; // dp
			float px = TypedValue.applyDimension(
				TypedValue.COMPLEX_UNIT_DIP,
				listItemHeight,
				getResources().getDisplayMetrics()
			);
			scrollXScale = px;
			scrollYScale = px;
		}

		setHighRefreshRate();

		accessManager = (AccessibilityManager)context.getSystemService(Context.ACCESSIBILITY_SERVICE);
		imm = (InputMethodManager)context.getSystemService(Context.INPUT_METHOD_SERVICE);
		nhandle = onCreateView(this);
		setFocusable(true);
		setFocusableInTouchMode(true);
		focusCallback = new View.OnFocusChangeListener() {
			@Override public void onFocusChange(View v, boolean focus) {
				GioView.this.onFocusChange(nhandle, focus);
			}
		};
		setOnFocusChangeListener(focusCallback);
		surfCallbacks = new SurfaceHolder.Callback() {
			@Override public void surfaceCreated(SurfaceHolder holder) {
				// Ignore; surfaceChanged is guaranteed to be called immediately after this.
			}
			@Override public void surfaceChanged(SurfaceHolder holder, int format, int width, int height) {
				onSurfaceChanged(nhandle, getHolder().getSurface());
			}
			@Override public void surfaceDestroyed(SurfaceHolder holder) {
				onSurfaceDestroyed(nhandle);
			}
		};
		getHolder().addCallback(surfCallbacks);
	}

	@Override public boolean onKeyDown(int keyCode, KeyEvent event) {
		if (nhandle != 0) {
			onKeyEvent(nhandle, keyCode, event.getUnicodeChar(), true, event.getEventTime());
		}
		return false;
	}

	@Override public boolean onKeyUp(int keyCode, KeyEvent event) {
		if (nhandle != 0) {
			onKeyEvent(nhandle, keyCode, event.getUnicodeChar(), false, event.getEventTime());
		}
		return false;
	}

	@Override public boolean onGenericMotionEvent(MotionEvent event) {
		dispatchMotionEvent(event);
		return true;
	}

	@Override public boolean onTouchEvent(MotionEvent event) {
		// Ask for unbuffered events. Flutter and Chrome do it
		// so assume it's good for us as well.
		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
			requestUnbufferedDispatch(event);
		}

		dispatchMotionEvent(event);
		return true;
	}

	private void setCursor(int id) {
		if (Build.VERSION.SDK_INT < Build.VERSION_CODES.N) {
			return;
		}
		PointerIcon pointerIcon = PointerIcon.getSystemIcon(getContext(), id);
		setPointerIcon(pointerIcon);
	}

	private void setOrientation(int id, int fallback) {
		if (Build.VERSION.SDK_INT < Build.VERSION_CODES.JELLY_BEAN_MR2) {
			id = fallback;
		}
		((Activity) this.getContext()).setRequestedOrientation(id);
	}

	private void setFullscreen(boolean enabled) {
		int flags = this.getSystemUiVisibility();
		if (enabled) {
		   flags |= SYSTEM_UI_FLAG_IMMERSIVE_STICKY;
		   flags |= SYSTEM_UI_FLAG_HIDE_NAVIGATION;
		   flags |= SYSTEM_UI_FLAG_FULLSCREEN;
		   flags |= SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN;
		} else {
		   flags &= ~SYSTEM_UI_FLAG_IMMERSIVE_STICKY;
		   flags &= ~SYSTEM_UI_FLAG_HIDE_NAVIGATION;
		   flags &= ~SYSTEM_UI_FLAG_FULLSCREEN;
		   flags &= ~SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN;
		}
		this.setSystemUiVisibility(flags);
	}

	private enum Bar {
		NAVIGATION,
		STATUS,
	}

	private void setBarColor(Bar t, int color, int luminance) {
		if (Build.VERSION.SDK_INT < Build.VERSION_CODES.LOLLIPOP) {
			return;
		}

		Window window = ((Activity) this.getContext()).getWindow();

		int insetsMask;
		int viewMask;

		switch (t) {
		case STATUS:
			insetsMask = WindowInsetsController.APPEARANCE_LIGHT_STATUS_BARS;
			viewMask = View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR;
			window.setStatusBarColor(color);
			break;
		case NAVIGATION:
			insetsMask = WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS;
			viewMask = View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR;
			window.setNavigationBarColor(color);
			break;
		default:
			throw new RuntimeException("invalid bar type");
		}

		if (Build.VERSION.SDK_INT < Build.VERSION_CODES.M) {
			return;
		}

		if (Build.VERSION.SDK_INT < Build.VERSION_CODES.R) {
			int flags = this.getSystemUiVisibility();
			if (luminance > 128) {
			   flags |=  viewMask;
			} else {
			   flags &= ~viewMask;
			}
			this.setSystemUiVisibility(flags);
			return;
		}

		WindowInsetsController insetsController = window.getInsetsController();
		if (insetsController == null) {
			return;
		}
		if (luminance > 128) {
			insetsController.setSystemBarsAppearance(insetsMask, insetsMask);
		} else {
			insetsController.setSystemBarsAppearance(0, insetsMask);
		}
	}

	private void setStatusColor(int color, int luminance) {
		this.setBarColor(Bar.STATUS, color, luminance);
	}

	private void setNavigationColor(int color, int luminance) {
		this.setBarColor(Bar.NAVIGATION, color, luminance);
	}

	private void setHighRefreshRate() {
		Context context = getContext();
		Display display = context.getDisplay();
		Display.Mode[] supportedModes = display.getSupportedModes();
		if (supportedModes.length <= 1) {
			// Nothing to set
			return;
		}

		Display.Mode currentMode = display.getMode();
		int currentWidth = currentMode.getPhysicalWidth();
		int currentHeight = currentMode.getPhysicalHeight();

		float minRefreshRate = -1;
		float maxRefreshRate = -1;
		float bestRefreshRate = -1;
		int bestModeId = -1;
		for (Display.Mode mode : supportedModes) {
			float refreshRate = mode.getRefreshRate();
			float width = mode.getPhysicalWidth();
			float height = mode.getPhysicalHeight();

			if (minRefreshRate == -1 || refreshRate < minRefreshRate) {
				minRefreshRate = refreshRate;
			}
			if (maxRefreshRate == -1 || refreshRate > maxRefreshRate) {
				maxRefreshRate = refreshRate;
			}

			boolean refreshRateIsBetter = bestRefreshRate == -1 || refreshRate > bestRefreshRate;
			if (width == currentWidth && height == currentHeight && refreshRateIsBetter) {
				int modeId = mode.getModeId();
				bestRefreshRate = refreshRate;
				bestModeId = modeId;
			}
		}

		if (bestModeId == -1) {
			// Not expecting this but just in case
			return;
		}

		if (minRefreshRate == maxRefreshRate) {
			// Can't improve the refresh rate
			return;
		}

		Window window = ((Activity) context).getWindow();
		WindowManager.LayoutParams layoutParams = window.getAttributes();
		layoutParams.preferredDisplayModeId = bestModeId;
		window.setAttributes(layoutParams);
	}

	@Override protected boolean dispatchHoverEvent(MotionEvent event) {
		if (!accessManager.isTouchExplorationEnabled()) {
			return super.dispatchHoverEvent(event);
		}
		switch (event.getAction()) {
		case MotionEvent.ACTION_HOVER_ENTER:
			// Fall through.
		case MotionEvent.ACTION_HOVER_MOVE:
			onTouchExploration(nhandle, event.getX(), event.getY());
			break;
		case MotionEvent.ACTION_HOVER_EXIT:
			onExitTouchExploration(nhandle);
			break;
		}
		return true;
	}

	void sendA11yEvent(int eventType, int viewId) {
		if (!accessManager.isEnabled()) {
			return;
		}
		AccessibilityEvent event = obtainA11yEvent(eventType, viewId);
		getParent().requestSendAccessibilityEvent(this, event);
	}

	AccessibilityEvent obtainA11yEvent(int eventType, int viewId) {
		AccessibilityEvent event = AccessibilityEvent.obtain(eventType);
		event.setPackageName(getContext().getPackageName());
		event.setSource(this, viewId);
		return event;
	}

	boolean isA11yActive() {
		return accessManager.isEnabled();
	}

	void sendA11yChange(int viewId) {
		if (!accessManager.isEnabled()) {
			return;
		}
		AccessibilityEvent event = obtainA11yEvent(AccessibilityEvent.TYPE_WINDOW_CONTENT_CHANGED, viewId);
		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
			event.setContentChangeTypes(AccessibilityEvent.CONTENT_CHANGE_TYPE_SUBTREE);
		}
		getParent().requestSendAccessibilityEvent(this, event);
	}

	private void dispatchMotionEvent(MotionEvent event) {
		if (nhandle == 0) {
			return;
		}
		for (int j = 0; j < event.getHistorySize(); j++) {
			long time = event.getHistoricalEventTime(j);
			for (int i = 0; i < event.getPointerCount(); i++) {
				onTouchEvent(
						nhandle,
						event.ACTION_MOVE,
						event.getPointerId(i),
						event.getToolType(i),
						event.getHistoricalX(i, j),
						event.getHistoricalY(i, j),
						scrollXScale*event.getHistoricalAxisValue(MotionEvent.AXIS_HSCROLL, i, j),
						scrollYScale*event.getHistoricalAxisValue(MotionEvent.AXIS_VSCROLL, i, j),
						event.getButtonState(),
						time);
			}
		}
		int act = event.getActionMasked();
		int idx = event.getActionIndex();
		for (int i = 0; i < event.getPointerCount(); i++) {
			int pact = event.ACTION_MOVE;
			if (i == idx) {
				pact = act;
			}
			onTouchEvent(
					nhandle,
					pact,
					event.getPointerId(i),
					event.getToolType(i),
					event.getX(i), event.getY(i),
					scrollXScale*event.getAxisValue(MotionEvent.AXIS_HSCROLL, i),
					scrollYScale*event.getAxisValue(MotionEvent.AXIS_VSCROLL, i),
					event.getButtonState(),
					event.getEventTime());
		}
	}

	@Override public InputConnection onCreateInputConnection(EditorInfo editor) {
		Snippet snip = getSnippet();
		editor.inputType = this.keyboardHint;
		editor.imeOptions = EditorInfo.IME_FLAG_NO_FULLSCREEN | EditorInfo.IME_FLAG_NO_EXTRACT_UI;
		editor.initialSelStart = imeToUTF16(nhandle, imeSelectionStart(nhandle));
		editor.initialSelEnd = imeToUTF16(nhandle, imeSelectionEnd(nhandle));
		int selStart = editor.initialSelStart - snip.offset;
		editor.initialCapsMode = TextUtils.getCapsMode(snip.snippet, selStart, this.keyboardHint);
		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
			editor.setInitialSurroundingSubText(snip.snippet, imeToUTF16(nhandle, snip.offset));
		}
		imeSetComposingRegion(nhandle, -1, -1);
		return new GioInputConnection();
	}

	void setInputHint(int hint) {
		if (hint == this.keyboardHint) {
			return;
		}
		this.keyboardHint = hint;
		restartInput();
	}

	void showTextInput() {
		GioView.this.requestFocus();
		imm.showSoftInput(GioView.this, 0);
	}

	void hideTextInput() {
		imm.hideSoftInputFromWindow(getWindowToken(), 0);
	}

	@Override protected boolean fitSystemWindows(Rect insets) {
		if (nhandle != 0) {
			onWindowInsets(nhandle, insets.top, insets.right, insets.bottom, insets.left);
		}
		return true;
	}

	void postFrameCallback() {
		Choreographer.getInstance().removeFrameCallback(this);
		Choreographer.getInstance().postFrameCallback(this);
	}

	@Override public void doFrame(long nanos) {
		if (nhandle != 0) {
			onFrameCallback(nhandle);
		}
	}

	int getDensity() {
		return getResources().getDisplayMetrics().densityDpi;
	}

	float getFontScale() {
		return getResources().getConfiguration().fontScale;
	}

	public void start() {
		if (nhandle != 0) {
			onStartView(nhandle);
		}
	}

	public void stop() {
		if (nhandle != 0) {
			onStopView(nhandle);
		}
	}

	public void destroy() {
		if (nhandle != 0) {
			onDestroyView(nhandle);
		}
	}

	protected void unregister() {
		setOnFocusChangeListener(null);
		getHolder().removeCallback(surfCallbacks);
		nhandle = 0;
	}

	public void configurationChanged() {
		if (nhandle != 0) {
			onConfigurationChanged(nhandle);
		}
	}

	public boolean backPressed() {
		if (nhandle == 0) {
			return false;
		}
		return onBack(nhandle);
	}

	void restartInput() {
		imm.restartInput(this);
	}

	void updateSelection() {
		int selStart = imeToUTF16(nhandle, imeSelectionStart(nhandle));
		int selEnd = imeToUTF16(nhandle, imeSelectionEnd(nhandle));
		int compStart = imeToUTF16(nhandle, imeComposingStart(nhandle));
		int compEnd = imeToUTF16(nhandle, imeComposingEnd(nhandle));
		imm.updateSelection(this, selStart, selEnd, compStart, compEnd);
	}

	void updateCaret(float m00, float m01, float m02, float m10, float m11, float m12, float caretX, float caretTop, float caretBase, float caretBottom) {
		if (Build.VERSION.SDK_INT < Build.VERSION_CODES.LOLLIPOP) {
			return;
		}
		Matrix m = new Matrix();
		m.setValues(new float[]{m00, m01, m02, m10, m11, m12, 0.0f, 0.0f, 1.0f});
		m.setConcat(getMatrix(), m);
		int selStart = imeSelectionStart(nhandle);
		int selEnd = imeSelectionEnd(nhandle);
		int compStart = imeComposingStart(nhandle);
		int compEnd = imeComposingEnd(nhandle);
		Snippet snip = getSnippet();
		String composing = "";
		if (compStart != -1) {
			composing = snip.substringRunes(compStart, compEnd);
		}
		CursorAnchorInfo inf = new CursorAnchorInfo.Builder()
			.setMatrix(m)
			.setComposingText(imeToUTF16(nhandle, compStart), composing)
			.setSelectionRange(imeToUTF16(nhandle, selStart), imeToUTF16(nhandle, selEnd))
			.setInsertionMarkerLocation(caretX, caretTop, caretBase, caretBottom, 0)
			.build();
		imm.updateCursorAnchorInfo(this, inf);
	}

	static private native long onCreateView(GioView view);
	static private native void onDestroyView(long handle);
	static private native void onStartView(long handle);
	static private native void onStopView(long handle);
	static private native void onSurfaceDestroyed(long handle);
	static private native void onSurfaceChanged(long handle, Surface surface);
	static private native void onConfigurationChanged(long handle);
	static private native void onWindowInsets(long handle, int top, int right, int bottom, int left);
	static public native void onLowMemory();
	static private native void onTouchEvent(long handle, int action, int pointerID, int tool, float x, float y, float scrollX, float scrollY, int buttons, long time);
	static private native void onKeyEvent(long handle, int code, int character, boolean pressed, long time);
	static private native void onFrameCallback(long handle);
	static private native boolean onBack(long handle);
	static private native void onFocusChange(long handle, boolean focus);
	static private native AccessibilityNodeInfo initializeAccessibilityNodeInfo(long handle, int viewId, int screenX, int screenY, AccessibilityNodeInfo info);
	static private native void onTouchExploration(long handle, float x, float y);
	static private native void onExitTouchExploration(long handle);
	static private native void onA11yFocus(long handle, int viewId);
	static private native void onClearA11yFocus(long handle, int viewId);
	static private native void imeSetSnippet(long handle, int start, int end);
	static private native String imeSnippet(long handle);
	static private native int imeSnippetStart(long handle);
	static private native int imeSelectionStart(long handle);
	static private native int imeSelectionEnd(long handle);
	static private native int imeComposingStart(long handle);
	static private native int imeComposingEnd(long handle);
	static private native int imeReplace(long handle, int start, int end, String text);
	static private native int imeSetSelection(long handle, int start, int end);
	static private native int imeSetComposingRegion(long handle, int start, int end);
	// imeToRunes converts the Java character index into runes (Java code points).
	static private native int imeToRunes(long handle, int chars);
	// imeToUTF16 converts the rune index into Java characters.
	static private native int imeToUTF16(long handle, int runes);

	private class GioInputConnection implements InputConnection {
		private int batchDepth;

		@Override public boolean beginBatchEdit() {
			batchDepth++;
			return true;
		}

		@Override public boolean endBatchEdit() {
			batchDepth--;
			return batchDepth > 0;
		}

		@Override public boolean clearMetaKeyStates(int states) {
			return false;
		}

		@Override public boolean commitCompletion(CompletionInfo text) {
			return false;
		}

		@Override public boolean commitCorrection(CorrectionInfo info) {
			return false;
		}

		@Override public boolean commitText(CharSequence text, int cursor) {
			setComposingText(text, cursor);
			return finishComposingText();
		}

		@Override public boolean deleteSurroundingText(int beforeChars, int afterChars) {
			// translate before and after to runes.
			int selStart = imeSelectionStart(nhandle);
			int selEnd = imeSelectionEnd(nhandle);
			int before = selStart - imeToRunes(nhandle, imeToUTF16(nhandle, selStart) - beforeChars);
			int after = selEnd - imeToRunes(nhandle, imeToUTF16(nhandle, selEnd) - afterChars);
			return deleteSurroundingTextInCodePoints(before, after);
		}

		@Override public boolean finishComposingText() {
			imeSetComposingRegion(nhandle, -1, -1);
			return true;
		}

		@Override public int getCursorCapsMode(int reqModes) {
			Snippet snip = getSnippet();
			int selStart = imeSelectionStart(nhandle);
			int off = imeToUTF16(nhandle, selStart - snip.offset);
			if (off < 0 || off > snip.snippet.length()) {
				return 0;
			}
			return TextUtils.getCapsMode(snip.snippet, off, reqModes);
		}

		@Override public ExtractedText getExtractedText(ExtractedTextRequest request, int flags) {
			return null;
		}

		@Override public CharSequence getSelectedText(int flags) {
			Snippet snip = getSnippet();
			int selStart = imeSelectionStart(nhandle);
			int selEnd = imeSelectionEnd(nhandle);
			String sub = snip.substringRunes(selStart, selEnd);
			return sub;
		}

		@Override public CharSequence getTextAfterCursor(int n, int flags) {
			Snippet snip = getSnippet();
			int selStart = imeSelectionStart(nhandle);
			int selEnd = imeSelectionEnd(nhandle);
			// n are in Java characters, but in worst case we'll just ask for more runes
			// than wanted.
			imeSetSnippet(nhandle, selStart - n, selEnd + n);
			int start = selEnd;
			int end = imeToRunes(nhandle, imeToUTF16(nhandle, selEnd) + n);
			String ret = snip.substringRunes(start, end);
			return ret;
		}

		@Override public CharSequence getTextBeforeCursor(int n, int flags) {
			Snippet snip = getSnippet();
			int selStart = imeSelectionStart(nhandle);
			int selEnd = imeSelectionEnd(nhandle);
			// n are in Java characters, but in worst case we'll just ask for more runes
			// than wanted.
			imeSetSnippet(nhandle, selStart - n, selEnd + n);
			int start = imeToRunes(nhandle, imeToUTF16(nhandle, selStart) - n);
			int end = selStart;
			String ret = snip.substringRunes(start, end);
			return ret;
		}

		@Override public boolean performContextMenuAction(int id) {
			return false;
		}

		@Override public boolean performEditorAction(int editorAction) {
			long eventTime = SystemClock.uptimeMillis();
			// Translate to enter key.
			onKeyEvent(nhandle, KeyEvent.KEYCODE_ENTER, '\n', true, eventTime);
			onKeyEvent(nhandle, KeyEvent.KEYCODE_ENTER, '\n', false, eventTime);
			return true;
		}

		@Override public boolean performPrivateCommand(String action, Bundle data) {
			return false;
		}

		@Override public boolean reportFullscreenMode(boolean enabled) {
			return false;
		}

		@Override public boolean sendKeyEvent(KeyEvent event) {
			boolean pressed = event.getAction() == KeyEvent.ACTION_DOWN;
			onKeyEvent(nhandle, event.getKeyCode(), event.getUnicodeChar(), pressed, event.getEventTime());
			return true;
		}

		@Override public boolean setComposingRegion(int startChars, int endChars) {
			int compStart = imeToRunes(nhandle, startChars);
			int compEnd = imeToRunes(nhandle, endChars);
			imeSetComposingRegion(nhandle, compStart, compEnd);
			return true;
		}

		@Override public boolean setComposingText(CharSequence text, int relCursor) {
			int start = imeComposingStart(nhandle);
			int end = imeComposingEnd(nhandle);
			if (start == -1 || end == -1) {
				start = imeSelectionStart(nhandle);
				end = imeSelectionEnd(nhandle);
			}
			String str = text.toString();
			imeReplace(nhandle, start, end, str);
			int cursor = start;
			int runes = str.codePointCount(0, str.length());
			if (relCursor > 0) {
				cursor += runes;
				relCursor--;
			}
			imeSetComposingRegion(nhandle, start, start + runes);

			// Move cursor.
			Snippet snip = getSnippet();
			cursor = imeToRunes(nhandle, imeToUTF16(nhandle, cursor) + relCursor);
			imeSetSelection(nhandle, cursor, cursor);
			return true;
		}

		@Override public boolean setSelection(int startChars, int endChars) {
			int start = imeToRunes(nhandle, startChars);
			int end = imeToRunes(nhandle, endChars);
			imeSetSelection(nhandle, start, end);
			return true;
		}

		/*@Override*/ public boolean requestCursorUpdates(int cursorUpdateMode) {
			// We always provide cursor updates.
			return true;
		}

		/*@Override*/ public void closeConnection() {
		}

		/*@Override*/ public Handler getHandler() {
			return null;
		}

		/*@Override*/ public boolean commitContent(InputContentInfo info, int flags, Bundle opts) {
			return false;
		}

		/*@Override*/ public boolean deleteSurroundingTextInCodePoints(int before, int after) {
			if (after > 0) {
				int selEnd = imeSelectionEnd(nhandle);
				imeReplace(nhandle, selEnd, selEnd + after, "");
			}
			if (before > 0) {
				int selStart = imeSelectionStart(nhandle);
				imeReplace(nhandle, selStart - before, selStart, "");
			}
			return true;
		}

		/*@Override*/ public SurroundingText getSurroundingText(int beforeChars, int afterChars, int flags) {
			Snippet snip = getSnippet();
			int selStart = imeSelectionStart(nhandle);
			int selEnd = imeSelectionEnd(nhandle);
			// Expanding in Java characters is ok.
			imeSetSnippet(nhandle, selStart - beforeChars, selEnd + afterChars);
			return new SurroundingText(snip.snippet, imeToUTF16(nhandle, selStart), imeToUTF16(nhandle, selEnd), imeToUTF16(nhandle, snip.offset));
		}
	}

	private Snippet getSnippet() {
		Snippet snip = new Snippet();
		snip.snippet = imeSnippet(nhandle);
		snip.offset = imeSnippetStart(nhandle);
		return snip;
	}

	// Snippet is like android.view.inputmethod.SurroundingText but available for Android < 31.
	private static class Snippet {
		String snippet;
		// offset of snippet into the entire editor content. It is in runes because we won't require
		// Gio editors to keep track of UTF-16 offsets. The distinction won't matter in practice because IMEs only
		// ever see snippets.
		int offset;

		// substringRunes returns the substring from start to end in runes. The resuls is
		// truncated to the snippet.
		String substringRunes(int start, int end) {
			start -= this.offset;
			end -= this.offset;
			int runes = snippet.codePointCount(0, snippet.length());
			if (start < 0) {
				start = 0;
			}
			if (end < 0) {
				end = 0;
			}
			if (start > runes) {
				start = runes;
			}
			if (end > runes) {
				end = runes;
			}
			return snippet.substring(
				snippet.offsetByCodePoints(0, start),
				snippet.offsetByCodePoints(0, end)
			);
		}
	}

	@Override public AccessibilityNodeProvider getAccessibilityNodeProvider() {
		return new AccessibilityNodeProvider() {
			private final int[] screenOff = new int[2];

			@Override public AccessibilityNodeInfo createAccessibilityNodeInfo(int viewId) {
				AccessibilityNodeInfo info = null;
				if (viewId == View.NO_ID) {
					info = AccessibilityNodeInfo.obtain(GioView.this);
					GioView.this.onInitializeAccessibilityNodeInfo(info);
				} else {
					info = AccessibilityNodeInfo.obtain(GioView.this, viewId);
					info.setPackageName(getContext().getPackageName());
					info.setVisibleToUser(true);
				}
				GioView.this.getLocationOnScreen(screenOff);
				info = GioView.this.initializeAccessibilityNodeInfo(nhandle, viewId, screenOff[0], screenOff[1], info);
				return info;
			}

			@Override public boolean performAction(int viewId, int action, Bundle arguments) {
				if (viewId == View.NO_ID) {
					return GioView.this.performAccessibilityAction(action, arguments);
				}
				switch (action) {
				case AccessibilityNodeInfo.ACTION_ACCESSIBILITY_FOCUS:
					GioView.this.onA11yFocus(nhandle, viewId);
					GioView.this.sendA11yEvent(AccessibilityEvent.TYPE_VIEW_ACCESSIBILITY_FOCUSED, viewId);
					return true;
				case AccessibilityNodeInfo.ACTION_CLEAR_ACCESSIBILITY_FOCUS:
					GioView.this.onClearA11yFocus(nhandle, viewId);
					GioView.this.sendA11yEvent(AccessibilityEvent.TYPE_VIEW_ACCESSIBILITY_FOCUS_CLEARED, viewId);
					return true;
				}
				return false;
			}
		};
	}
}
