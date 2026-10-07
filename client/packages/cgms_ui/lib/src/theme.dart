import 'package:flutter/material.dart';

/// Opaque Cgmsart reading surfaces and semantic accents.
abstract final class CgmsColors {
  static const canvas = Color(0xFF101322);
  static const surface = Color(0xFF1D2335);
  static const raised = Color(0xFF2B344B);
  static const text = Color(0xFFF5F1EA);
  static const muted = Color(0xFFBAC4D6);
  static const action = Color(0xFFFF795E);
  static const cyan = Color(0xFF69D8E4);
  static const lime = Color(0xFFDFFF7A);
  static const border = Color(0xFF8595B0);
  static const heart = Color(0xFFF5A4BE);
  static const diamond = Color(0xFFFFD275);
  static const club = Color(0xFF83DCD3);
  static const spade = Color(0xFFB6B2FF);
  static const divider = Color(0xFF414B65);
  static const error = Color(0xFFFFA1B7);
}

/// Shared spacing for controls, related content and reading surfaces.
abstract final class CgmsSpacing {
  static const controlGap = 8.0;
  static const sectionGap = 16.0;
  static const panelInset = 16.0;
  static const pageInset = 24.0;
}

abstract final class CgmsShape {
  static const control = 8.0;
  static const panel = 12.0;
}

abstract final class CgmsTheme {
  static ThemeData dark() {
    const body = 'packages/cgms_ui/Atkinson Hyperlegible Next';
    const display = 'packages/cgms_ui/Barlow Condensed';
    final textTheme = ThemeData.dark().textTheme
        .apply(fontFamily: body)
        .copyWith(
          displayLarge: const TextStyle(
            fontFamily: display,
            fontSize: 56,
            fontWeight: FontWeight.w700,
            height: 1.05,
          ),
          displayMedium: const TextStyle(
            fontFamily: display,
            fontSize: 44,
            fontWeight: FontWeight.w700,
            height: 1.08,
          ),
          displaySmall: const TextStyle(
            fontFamily: display,
            fontSize: 36,
            fontWeight: FontWeight.w700,
            height: 1.1,
          ),
          headlineLarge: const TextStyle(
            fontFamily: display,
            fontSize: 36,
            fontWeight: FontWeight.w700,
            height: 1.1,
          ),
          headlineMedium: const TextStyle(
            fontFamily: display,
            fontSize: 30,
            fontWeight: FontWeight.w700,
            height: 1.15,
          ),
          headlineSmall: const TextStyle(
            fontFamily: display,
            fontSize: 26,
            fontWeight: FontWeight.w700,
            height: 1.15,
          ),
          titleLarge: const TextStyle(
            fontFamily: body,
            fontSize: 20,
            fontWeight: FontWeight.w700,
            height: 1.3,
          ),
          titleMedium: const TextStyle(
            fontFamily: body,
            fontSize: 16,
            fontWeight: FontWeight.w700,
            height: 1.4,
          ),
          bodyLarge: const TextStyle(
            fontFamily: body,
            fontSize: 16,
            height: 1.45,
          ),
          bodyMedium: const TextStyle(
            fontFamily: body,
            fontSize: 15,
            height: 1.45,
          ),
          bodySmall: const TextStyle(
            fontFamily: body,
            fontSize: 14,
            height: 1.4,
          ),
          labelLarge: const TextStyle(
            fontFamily: body,
            fontSize: 16,
            fontWeight: FontWeight.w700,
            height: 1.3,
          ),
          labelMedium: const TextStyle(
            fontFamily: body,
            fontSize: 14,
            fontWeight: FontWeight.w700,
            height: 1.3,
          ),
          labelSmall: const TextStyle(
            fontFamily: body,
            fontSize: 14,
            height: 1.3,
          ),
        )
        .apply(
          bodyColor: CgmsColors.text,
          displayColor: CgmsColors.text,
          fontFamilyFallback: const ['packages/cgms_ui/Noto Sans Symbols2'],
        );
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(CgmsShape.control),
    );
    final focusSide = WidgetStateProperty.resolveWith<BorderSide>(
      (states) => BorderSide(
        color: states.contains(WidgetState.focused)
            ? CgmsColors.lime
            : CgmsColors.border,
        width: states.contains(WidgetState.focused) ? 3 : 1,
      ),
    );
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      scaffoldBackgroundColor: CgmsColors.canvas,
      fontFamily: body,
      textTheme: textTheme,
      colorScheme: const ColorScheme.dark(
        primary: CgmsColors.action,
        onPrimary: CgmsColors.canvas,
        secondary: CgmsColors.cyan,
        onSecondary: CgmsColors.canvas,
        surface: CgmsColors.surface,
        onSurface: CgmsColors.text,
        onSurfaceVariant: CgmsColors.muted,
        outline: CgmsColors.border,
        error: CgmsColors.error,
        onError: CgmsColors.canvas,
      ),
      dividerColor: CgmsColors.divider,
      focusColor: CgmsColors.lime.withValues(alpha: 0.18),
      hoverColor: CgmsColors.cyan.withValues(alpha: 0.09),
      disabledColor: CgmsColors.muted,
      cardTheme: CardThemeData(
        color: CgmsColors.surface,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(CgmsShape.panel),
          side: const BorderSide(color: CgmsColors.divider),
        ),
      ),
      chipTheme: ChipThemeData(
        backgroundColor: CgmsColors.surface,
        selectedColor: CgmsColors.raised,
        disabledColor: CgmsColors.canvas,
        checkmarkColor: CgmsColors.cyan,
        labelStyle: textTheme.labelMedium,
        secondaryLabelStyle: textTheme.labelMedium?.copyWith(
          color: CgmsColors.cyan,
        ),
        shape: shape,
        side: const BorderSide(color: CgmsColors.border),
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 10),
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: CgmsColors.surface,
        surfaceTintColor: Colors.transparent,
        textStyle: textTheme.bodyLarge,
        shape: shape,
      ),
      navigationRailTheme: NavigationRailThemeData(
        backgroundColor: CgmsColors.canvas,
        indicatorColor: CgmsColors.raised,
        indicatorShape: shape,
        selectedIconTheme: const IconThemeData(color: CgmsColors.cyan),
        unselectedIconTheme: const IconThemeData(color: CgmsColors.muted),
        selectedLabelTextStyle: textTheme.labelMedium?.copyWith(
          color: CgmsColors.cyan,
        ),
        unselectedLabelTextStyle: textTheme.labelMedium?.copyWith(
          color: CgmsColors.muted,
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: ButtonStyle(
          minimumSize: const WidgetStatePropertyAll(Size(48, 48)),
          padding: const WidgetStatePropertyAll(
            EdgeInsets.symmetric(horizontal: 20, vertical: 14),
          ),
          shape: WidgetStatePropertyAll(shape),
          side: WidgetStateProperty.resolveWith(
            (states) => BorderSide(
              color: states.contains(WidgetState.focused)
                  ? CgmsColors.lime
                  : Colors.transparent,
              width: 3,
            ),
          ),
          backgroundColor: WidgetStateProperty.resolveWith(
            (states) => states.contains(WidgetState.disabled)
                ? CgmsColors.raised
                : CgmsColors.action,
          ),
          foregroundColor: WidgetStateProperty.resolveWith(
            (states) => states.contains(WidgetState.disabled)
                ? CgmsColors.muted
                : CgmsColors.canvas,
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: ButtonStyle(
          minimumSize: const WidgetStatePropertyAll(Size(48, 48)),
          padding: const WidgetStatePropertyAll(
            EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          ),
          foregroundColor: const WidgetStatePropertyAll(CgmsColors.text),
          shape: WidgetStatePropertyAll(shape),
          side: focusSide,
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: ButtonStyle(
          minimumSize: const WidgetStatePropertyAll(Size(48, 48)),
          foregroundColor: const WidgetStatePropertyAll(CgmsColors.cyan),
          shape: WidgetStatePropertyAll(shape),
          side: WidgetStateProperty.resolveWith(
            (states) => BorderSide(
              color: states.contains(WidgetState.focused)
                  ? CgmsColors.lime
                  : Colors.transparent,
              width: 3,
            ),
          ),
        ),
      ),
      iconButtonTheme: IconButtonThemeData(
        style: ButtonStyle(
          minimumSize: const WidgetStatePropertyAll(Size(48, 48)),
          foregroundColor: const WidgetStatePropertyAll(CgmsColors.text),
          side: WidgetStateProperty.resolveWith(
            (states) => BorderSide(
              color: states.contains(WidgetState.focused)
                  ? CgmsColors.lime
                  : Colors.transparent,
              width: 3,
            ),
          ),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: CgmsColors.surface,
        labelStyle: const TextStyle(color: CgmsColors.muted),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(CgmsShape.control),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(CgmsShape.control),
          borderSide: const BorderSide(color: CgmsColors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(CgmsShape.control),
          borderSide: const BorderSide(color: CgmsColors.lime, width: 3),
        ),
        contentPadding: const EdgeInsets.all(16),
      ),
      tooltipTheme: const TooltipThemeData(
        textStyle: TextStyle(
          color: CgmsColors.canvas,
          fontSize: 14,
          fontFamily: body,
        ),
        decoration: BoxDecoration(
          color: CgmsColors.text,
          borderRadius: BorderRadius.all(Radius.circular(4)),
        ),
      ),
      scrollbarTheme: ScrollbarThemeData(
        thumbVisibility: const WidgetStatePropertyAll(true),
        thumbColor: WidgetStatePropertyAll(
          CgmsColors.border.withValues(alpha: 0.7),
        ),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: CgmsColors.surface,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(CgmsShape.panel),
          side: const BorderSide(color: CgmsColors.border),
        ),
        titleTextStyle: textTheme.headlineSmall,
      ),
      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: CgmsColors.surface,
        surfaceTintColor: Colors.transparent,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: CgmsColors.canvas,
        foregroundColor: CgmsColors.text,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
      ),
      textSelectionTheme: const TextSelectionThemeData(
        cursorColor: CgmsColors.cyan,
        selectionColor: Color(0x8069D8E4),
        selectionHandleColor: CgmsColors.cyan,
      ),
    );
  }
}
