// SPDX-License-Identifier: Unlicense OR MIT

package org.gioui;

import android.app.Activity;
import android.os.Bundle;
import android.content.res.Configuration;

public final class GioActivity extends Activity {
	private GioView view;

	@Override public void onCreate(Bundle state) {
		super.onCreate(state);

		this.view = new GioView(this);

		setContentView(view);
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
