// SPDX-License-Identifier: Unlicense OR MIT

package org.gioui;

import android.app.Activity;
import android.os.Bundle;
import android.content.res.Configuration;
import android.view.ViewGroup;
import android.view.View;
import android.view.ViewGroup;
import android.widget.FrameLayout;

public final class GioActivity extends Activity {
	private GioView view;
	public FrameLayout layer;

	@Override public void onCreate(Bundle state) {
            super.onCreate(state);

            layer = new FrameLayout(this);
            view = new GioView(this);

            view.setLayoutParams(new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT
            ));
            view.setFocusable(true);
            view.setFocusableInTouchMode(true);

            layer.addView(view);
            setContentView(layer);
	}

	@Override public void onDestroy() {
		view.destroy();
		super.onDestroy();
	}

	@Override public void onStart() {
		super.onStart();
		view.start();
	}

	@Override public void onStop() {
		view.stop();
		super.onStop();
	}

	@Override public void onConfigurationChanged(Configuration c) {
		super.onConfigurationChanged(c);
		view.configurationChanged();
	}

	@Override public void onLowMemory() {
		super.onLowMemory();
		GioView.onLowMemory();
	}

	@Override public void onBackPressed() {
		if (!view.backPressed())
			super.onBackPressed();
	}
}
