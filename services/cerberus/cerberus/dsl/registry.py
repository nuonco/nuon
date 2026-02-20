"""Canary registry and @canary decorator.

Usage:
    @canary(name="basic_deploy", description="...", category="basic")
    class BasicDeploy:
        app = App(...)
        installs = [Install(...)]
        steps = [deploy(), verify_status("healthy")]
"""

from __future__ import annotations

import importlib
import pkgutil
from dataclasses import dataclass, field
from datetime import timedelta
from typing import Any

from cerberus.dsl.base import App, Install
from cerberus.dsl.primitives import Step


@dataclass
class CanaryDefinition:
    """Metadata about a registered canary class."""

    name: str
    description: str
    category: str
    timeout: timedelta
    cls: type
    app: App
    installs: list[Install]
    steps: list[Step]


# Global registry: name -> CanaryDefinition
_registry: dict[str, CanaryDefinition] = {}


def canary(
    name: str,
    *,
    description: str = "",
    category: str = "general",
    timeout: timedelta = timedelta(hours=24),
):
    """Class decorator that registers a canary definition."""

    def decorator(cls):
        app = getattr(cls, "app", None)
        if app is None or not isinstance(app, App):
            raise TypeError(f"Canary {name} must declare an 'app' class attribute of type App")

        installs = getattr(cls, "installs", [])
        if not isinstance(installs, list):
            raise TypeError(f"Canary {name} 'installs' must be a list of Install")

        steps = getattr(cls, "steps", [])
        if not isinstance(steps, list):
            raise TypeError(f"Canary {name} 'steps' must be a list of Step")

        defn = CanaryDefinition(
            name=name,
            description=description,
            category=category,
            timeout=timeout,
            cls=cls,
            app=app,
            installs=installs,
            steps=steps,
        )

        _registry[name] = defn

        # Stash definition on the class for introspection
        cls._cerberus_definition = defn

        return cls

    return decorator


def get_canary(name: str) -> CanaryDefinition:
    """Look up a canary by name. Raises KeyError if not found."""
    if name not in _registry:
        raise KeyError(f"Canary '{name}' not found. Available: {list(_registry.keys())}")
    return _registry[name]


def list_canaries() -> list[CanaryDefinition]:
    """Return all registered canary definitions."""
    return list(_registry.values())


def discover_canaries(package_name: str = "cerberus.canaries") -> None:
    """Import all modules in the canaries package to trigger registration."""
    try:
        package = importlib.import_module(package_name)
    except ImportError:
        return

    if not hasattr(package, "__path__"):
        return

    for _importer, module_name, _ispkg in pkgutil.iter_modules(package.__path__):
        importlib.import_module(f"{package_name}.{module_name}")
